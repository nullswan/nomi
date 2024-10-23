package transcription

import (
	"sync"
	"time"

	"github.com/nullswan/nomi/internal/audio"
)

const flushChanSz = 100

type AudioChunk struct {
	Data          []byte
	StartDuration time.Duration
	EndDuration   time.Duration
}

// BufferManager defines the methods for buffer managers.
type BufferManager interface {
	AddAudio(audioData []byte)
	GetAudio() (AudioChunk, bool)

	IsEmpty() bool

	Flush()
	Reset()
	Close()
}

type bufferManager struct {
	buffer            []byte
	overlapDuration   time.Duration
	minBufferDuration time.Duration
	overlapBytes      int

	sampleRate     int
	channels       int
	bytesPerSample int
	bitsPerSample  int

	mu         sync.Mutex
	flushChan  chan AudioChunk
	baseOffset time.Duration
}

func NewBufferManager(audioOpts *audio.AudioOptions) *bufferManager {
	return &bufferManager{
		flushChan:      make(chan AudioChunk, flushChanSz),
		sampleRate:     int(audioOpts.SampleRate),
		channels:       audioOpts.Channels,
		bytesPerSample: audioOpts.BytesPerSample,
		bitsPerSample:  audioOpts.BitsPerSample,
		baseOffset:     0,
	}
}

func (bm *bufferManager) SetMinBufferDuration(duration time.Duration) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.minBufferDuration = duration
}

func (bm *bufferManager) SetOverlapDuration(duration time.Duration) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.overlapDuration = duration
	bm.overlapBytes = bm.computeBytes(duration)
}

func (bm *bufferManager) AddAudio(audioData []byte) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.buffer = append(bm.buffer, audioData...)
	audioDuration := bm.computeDuration(len(bm.buffer))

	if audioDuration >= bm.minBufferDuration {
		bm.flushBuffer(audioDuration)
	}
}

func (bm *bufferManager) Flush() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.flushBuffer(0)
}

func (bm *bufferManager) flushBuffer(bufferDuration time.Duration) {
	if len(bm.buffer) == 0 {
		return
	}

	var duration time.Duration
	if bufferDuration > 0 {
		duration = bufferDuration
	} else {
		duration = bm.computeDuration(len(bm.buffer))
	}

	start := bm.baseOffset
	end := bm.baseOffset + duration

	chunk := AudioChunk{
		Data:          append([]byte{}, bm.buffer...),
		StartDuration: start,
		EndDuration:   end,
	}

	select {
	case bm.flushChan <- chunk:
		if bm.overlapBytes < len(bm.buffer) {
			bm.buffer = bm.buffer[len(bm.buffer)-bm.overlapBytes:]
			bm.baseOffset = end - bm.overlapDuration
		} else {
			bm.buffer = bm.buffer[:0]
			bm.baseOffset = end
		}
	default:
	}
}

func (bm *bufferManager) computeDuration(bufferLength int) time.Duration {
	if bm.sampleRate == 0 || bm.channels == 0 || bm.bytesPerSample == 0 {
		return 0
	}

	bytesPerSecond := bm.sampleRate * bm.channels * bm.bytesPerSample
	return time.Duration(
		bufferLength,
	) * time.Second / time.Duration(
		bytesPerSecond,
	)
}

func (bm *bufferManager) computeBytes(duration time.Duration) int {
	if bm.sampleRate == 0 || bm.channels == 0 || bm.bytesPerSample == 0 {
		return 0
	}

	bytesPerSecond := bm.sampleRate * bm.channels * bm.bytesPerSample
	return int(duration.Seconds() * float64(bytesPerSecond))
}

func (bm *bufferManager) GetAudio() (AudioChunk, bool) {
	audio, ok := <-bm.flushChan
	return audio, ok
}

func (bm *bufferManager) Reset() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.baseOffset = 0
	bm.buffer = bm.buffer[:0]
}

func (bm *bufferManager) Close() {
	bm.mu.Lock()
	bm.flushBuffer(0)
	bm.mu.Unlock()

	close(bm.flushChan)
}

func (bm *bufferManager) IsEmpty() bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	return len(bm.buffer) <= bm.overlapBytes
}

type simpleBufferManager struct {
	buffer     []byte
	sampleRate int
	channels   int
	mu         sync.Mutex
	flushChan  chan AudioChunk

	minBufferDuration time.Duration
}

func NewSimpleBufferManager(
	audioOpts *audio.AudioOptions,
) *simpleBufferManager {
	return &simpleBufferManager{
		flushChan:         make(chan AudioChunk, flushChanSz),
		sampleRate:        int(audioOpts.SampleRate),
		channels:          audioOpts.Channels,
		minBufferDuration: 0,
	}
}

func (sbm *simpleBufferManager) AddAudio(audioData []byte) {
	sbm.mu.Lock()
	defer sbm.mu.Unlock()

	sbm.buffer = append(sbm.buffer, audioData...)
}

func (sbm *simpleBufferManager) Flush() {
	sbm.mu.Lock()
	defer sbm.mu.Unlock()

	if len(sbm.buffer) == 0 {
		return
	}

	duration := sbm.computeDuration(len(sbm.buffer))
	if duration < sbm.minBufferDuration {
		// Drop the buffer if it's too short
		sbm.buffer = sbm.buffer[:0]
		return
	}

	chunk := AudioChunk{
		Data:          append([]byte{}, sbm.buffer...),
		StartDuration: 0,
		EndDuration:   duration,
	}

	select {
	case sbm.flushChan <- chunk:
		sbm.buffer = sbm.buffer[:0]
	default:
	}
}

func (sbm *simpleBufferManager) computeDuration(
	bufferLength int,
) time.Duration {
	if sbm.sampleRate == 0 || sbm.channels == 0 {
		return 0
	}

	bytesPerSecond := sbm.sampleRate * sbm.channels * 2 // assuming 16 bits per sample
	return time.Duration(
		bufferLength,
	) * time.Second / time.Duration(
		bytesPerSecond,
	)
}

func (sbm *simpleBufferManager) GetAudio() (AudioChunk, bool) {
	audio, ok := <-sbm.flushChan
	return audio, ok
}

func (sbm *simpleBufferManager) Reset() {
	sbm.mu.Lock()
	defer sbm.mu.Unlock()

	sbm.buffer = sbm.buffer[:0]
}

func (sbm *simpleBufferManager) Close() {
	sbm.mu.Lock()
	if len(sbm.buffer) > 0 {
		sbm.Flush()
	}
	sbm.mu.Unlock()
	close(sbm.flushChan)
}

func (sbm *simpleBufferManager) IsEmpty() bool {
	sbm.mu.Lock()
	defer sbm.mu.Unlock()

	return len(sbm.buffer) == 0
}

func (sbn *simpleBufferManager) SetMinBufferDuration(duration time.Duration) {
	sbn.mu.Lock()
	defer sbn.mu.Unlock()

	sbn.minBufferDuration = duration
}
