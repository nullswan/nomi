package voiceinput

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate      = 32000
	channels        = 1
	framesPerBuffer = 640
	bytesPerSample  = 2

	compressedChunkSize  = 20
	compressedMinChunks  = 6
	compressedBufferSize = framesPerBuffer * bytesPerSample * channels * compressedChunkSize

	stopWord  = "[STOP]"
	startWord = "[START]"
)

var (
	stream      *portaudio.Stream
	initOnce    sync.Once
	inputBuffer = make([]int16, framesPerBuffer*channels)
)

// InitializeAudio sets up the PortAudio stream
func InitializeAudio() {
	fmt.Println("Initializing audio...")

	initOnce.Do(func() {
		if err := portaudio.Initialize(); err != nil {
			log.Fatalf("Failed to initialize PortAudio: %v", err)
		}

		var err error
		stream, err = portaudio.OpenDefaultStream(
			channels,            // Input channels
			0,                   // Output channels
			float64(sampleRate), // Sample rate
			framesPerBuffer,     // Frames per buffer
			inputBuffer,         // Input buffer
		)
		if err != nil {
			log.Fatalf("Failed to open default stream: %v", err)
		}

		if err := stream.Start(); err != nil {
			log.Fatalf("Failed to start stream: %v", err)
		}
	})
}

// CleanUpAudio stops and closes the PortAudio stream
func CleanUpAudio() {
	if stream != nil {
		if err := stream.Stop(); err != nil {
			log.Printf("Failed to stop stream: %v", err)
		}
		if err := stream.Close(); err != nil {
			log.Printf("Failed to close stream: %v", err)
		}
		stream = nil
	}
	portaudio.Terminate()
}

// ReadAudioChunk reads a chunk of audio data from the microphone
func ReadAudioChunk(buffer []byte) (int, error) {
	if stream == nil {
		return 0, io.EOF
	}

	// Read data into the inputBuffer
	if err := stream.Read(); err != nil {
		return 0, fmt.Errorf("Failed to read from stream: %v", err)
	}

	// Convert int16 samples to bytes
	for i, sample := range inputBuffer {
		buffer[bytesPerSample*i] = byte(sample)
		buffer[bytesPerSample*i+1] = byte(sample >> 8)
	}

	return framesPerBuffer * bytesPerSample, nil
}
