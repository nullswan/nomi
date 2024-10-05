package voiceinput

import (
	"context"
	"io"
	"log"
)

// Run starts the voice input loop
// This MUST be run in a separate goroutine
func Run(ctx context.Context, ch chan<- []byte) {
	InitializeAudio()
	defer CleanUpAudio()

	audioChan := make(chan []byte)
	defer close(audioChan)

	// Goroutine to read audio and send to audioChan
	go func() {
		buffer := make(
			[]byte,
			framesPerBuffer*bytesPerSample,
		)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := ReadAudioChunk(buffer)
				if err != nil && err != io.EOF {
					log.Println("ReadAudioChunk error:", err)
					continue
				}
				if n == 0 {
					continue
				}

				// Make a copy of the buffer to avoid data race
				chunk := make([]byte, n)
				copy(chunk, buffer[:n])
				audioChan <- chunk
			}
		}
	}()

	vad := initializeVAD()
	speaking := false

	chunckAgg := make(
		[]byte,
		0,
		compressedBufferSize,
	)

	for {
		select {
		case <-ctx.Done():
			return
		case chunk, ok := <-audioChan:
			if !ok {
				return
			}
			// fmt.Printf("Ch: %v\n", chunk)
			if isSpeech(vad, chunk) {
				if !speaking {
					ch <- []byte(startWord)
					speaking = true
				}
				chunckAgg = append(chunckAgg, chunk...)
				if len(chunckAgg) == compressedChunkSize {
					flushAudioChunk(ch, &chunckAgg)
				}
			} else {
				if speaking {
					ch <- []byte(stopWord)
					speaking = false

					flushAudioChunk(ch, &chunckAgg)
				}
			}
		}
	}
}

func flushAudioChunk(ch chan<- []byte, chunks *[]byte) {
	if len(
		*chunks,
	) <= compressedBufferSize/compressedChunkSize*compressedMinChunks {
		*chunks = make(
			[]byte,
			0,
			compressedBufferSize,
		)
		return
	}

	// Flush the chunks
	ch <- compressAudio(*chunks)

	// Clear the chunks
	*chunks = make(
		[]byte,
		0,
		compressedBufferSize,
	)
	return
}
