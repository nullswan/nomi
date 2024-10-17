import (
	"sync"
	"time"

	"github.com/nullswan/nomi/internal/audio"
)

// AudioChunk represents a chunk of audio data with a timestamp and buffer source identifier.
type AudioChunk struct {
	Data         []byte
	Timestamp    time.Time
	BufferSource string // "primary" or "secondary"
}

// BufferManager manages audio buffering and flushing based on minimum buffer duration.
// It handles overlapping buffers to ensure continuity between chunks.
type BufferManager struct {
	buffer            []byte
	overlapDuration   time.Duration
	minBufferDuration time.Duration

	sampleRate     int
	channels       int
	bytesPerSample int
	bitsPerSample  int

	mu        sync.Mutex
	flushChan chan AudioChunk
	quitChan  chan struct{}

	bufferSource string // Identifier for the buffer instance ("primary" or "secondary")
}

// NewBufferManager creates a new BufferManager instance with the specified source identifier.
func NewBufferManager(audioOpts *audio.AudioOptions, bufferSource string) *BufferManager {
	bm := &BufferManager{
		flushChan:      make(chan AudioChunk, 2), // Buffer size accommodates two instances
		quitChan:       make(chan struct{}),
		sampleRate:     int(audioOpts.SampleRate),
		channels:       audioOpts.Channels,
		bytesPerSample: audioOpts.BytesPerSample,
		bitsPerSample:  audioOpts.BitsPerSample,
		bufferSource:   bufferSource,
	}
	go bm.run()
	return bm
}

// SetMinBufferDuration sets the minimum buffer duration before flushing.
func (bm *BufferManager) SetMinBufferDuration(duration time.Duration) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.minBufferDuration = duration
}

// SetOverlapDuration sets the duration of audio overlap between chunks.
func (bm *BufferManager) SetOverlapDuration(duration time.Duration) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.overlapDuration = duration
}

// AddAudio appends audio data to the buffer and flushes if the minimum duration is met.
func (bm *BufferManager) AddAudio(audio []byte) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.buffer = append(bm.buffer, audio...)
	audioDuration := bm.computeDuration(len(bm.buffer))

	if audioDuration >= bm.minBufferDuration {
		select {
		case bm.flushChan <- AudioChunk{
			Data:         append([]byte{}, bm.buffer...),
			Timestamp:    time.Now(),
			BufferSource: bm.bufferSource,
		}:
			overlapBytes := bm.computeBytes(bm.overlapDuration)
			if overlapBytes < len(bm.buffer) {
				bm.buffer = bm.buffer[len(bm.buffer)-overlapBytes:]
			} else {
				bm.buffer = bm.buffer[:0]
			}
		default:
			// Flush channel is full; skip flushing to avoid blocking
		}
	}
}

// Flush forces the buffer to flush its current data.
func (bm *BufferManager) Flush() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if len(bm.buffer) > 0 {
		select {
		case bm.flushChan <- AudioChunk{
			Data:         append([]byte{}, bm.buffer...),
			Timestamp:    time.Now(),
			BufferSource: bm.bufferSource,
		}:
			overlapBytes := bm.computeBytes(bm.overlapDuration)
			if overlapBytes < len(bm.buffer) {
				bm.buffer = bm.buffer[len(bm.buffer)-overlapBytes:]
			} else {
				bm.buffer = bm.buffer[:0]
			}
		default:
			// Flush channel is full; skip flushing to avoid blocking
		}
	}
}

// run starts the buffer manager's ticker for periodic checking and flushing.
func (bm *BufferManager) run() {
	ticker := time.NewTicker(500 * time.Millisecond) // Adjust tick duration as needed
	defer ticker.Stop()
	for {
		select {
		case <-bm.quitChan:
			return
		case <-ticker.C:
			bm.mu.Lock()
			bufferDuration := bm.computeDuration(len(bm.buffer))
			if bufferDuration >= bm.minBufferDuration {
				bm.mu.Unlock()
				bm.Flush()
			} else {
				bm.mu.Unlock()
			}
		}
	}
}

// computeDuration calculates the duration of the buffered audio.
func (bm *BufferManager) computeDuration(bufferLength int) time.Duration {
	if bm.sampleRate == 0 || bm.channels == 0 || bm.bytesPerSample == 0 {
		return 0
	}
	bytesPerSecond := bm.sampleRate * bm.channels * bm.bytesPerSample
	return time.Duration(bufferLength) * time.Second / time.Duration(bytesPerSecond)
}

// computeBytes calculates the number of bytes for a given duration.
func (bm *BufferManager) computeBytes(duration time.Duration) int {
	if bm.sampleRate == 0 || bm.channels == 0 || bm.bytesPerSample == 0 {
		return 0
	}
	bytesPerSecond := bm.sampleRate * bm.channels * bm.bytesPerSample
	return int(duration.Seconds() * float64(bytesPerSecond))
}

// GetAudio retrieves the next audio chunk from the flush channel.
func (bm *BufferManager) GetAudio() (AudioChunk, bool) {
	audio, ok := <-bm.flushChan
	return audio, ok
}

// Close gracefully shuts down the BufferManager by flushing remaining data and closing channels.
func (bm *BufferManager) Close() {
	bm.mu.Lock()
	// Flush remaining buffer if it exists
	if len(bm.buffer) > 0 {
		select {
		case bm.flushChan <- AudioChunk{
			Data:         append([]byte{}, bm.buffer...),
			Timestamp:    time.Now(),
			BufferSource: bm.bufferSource,
		}:
			bm.buffer = bm.buffer[:0]
		default:
			// Flush channel is full; skip flushing to avoid blocking
		}
	}
	bm.mu.Unlock()

	close(bm.flushChan)
	close(bm.quitChan)
}