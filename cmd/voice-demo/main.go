package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/nullswan/nomi/internal/audio"
	"github.com/nullswan/nomi/internal/logger"
	"github.com/nullswan/nomi/internal/transcription"
)

func main() {
	// Initialize logger
	logger := logger.Init()

	oaiKey := os.Getenv("OPENAI_API_KEY")
	if oaiKey == "" {
		logger.Error("OPENAI_API_KEY is not set")
		return
	}

	// Initialize PortAudio
	if err := portaudio.Initialize(); err != nil {
		logger.Error("Failed to initialize PortAudio", "error", err)
		return
	}
	defer portaudio.Terminate()

	audioOpts := &audio.AudioOptions{}
	audioOpts, err := audio.ComputeAudioOptions(audioOpts)
	if err != nil {
		logger.
			With("error", err).
			Error("Failed to compute audio options")
		return
	}

	callback := func(text string) {
		fmt.Println("Transcribed Text:", text)
	}

	ts := transcription.NewTranscriptionServer(
		oaiKey,
		logger,
		callback,
		audioOpts,
		transcription.WithOverlapDuration(500*time.Millisecond),
		transcription.WithMinChunkDuration(500*time.Millisecond),
		transcription.WithMaxChunkDuration(2000*time.Millisecond),
		transcription.WithContextTimeout(30*time.Second),
		transcription.WithEnableFixing(false),
	)
	defer ts.Close()

	// Initialize VAD
	vad := audio.NewVAD(
		audio.VADConfig{
			EnergyThreshold: 0.02,                   // Adjust based on testing
			FlushInterval:   250 * time.Millisecond, // Flush every 250ms
			SilenceDuration: 800 * time.Millisecond, // Silence duration to detect end of speech
			PauseDuration:   200 * time.Millisecond, // Pause duration to detect end of speech
		},
		audio.VADCallbacks{
			OnSpeechStart: func() {
				logger.Info("VAD: Speech started")
				// ts.Reset()
			},
			OnSpeechEnd: func() {
				logger.Info("VAD: Speech ended")

				ts.Flush()
				final := ts.GetText()

				fmt.Println("Final Text:", final)
			},
			OnFlush: func(buffer []float32) {
				logger.With("buf_sz", len(buffer)).
					Info("VAD: Buffer flushed")

				data, err := audio.Float32ToPCM(buffer)
				if err != nil {
					logger.Error(
						"Failed to convert float32 to PCM",
						"error",
						err,
					)
					return
				}

				ts.AddAudio(data)
			},
			OnPause: func() {
				logger.Info("VAD: Speech paused")
				ts.Flush()
			},
		},
		logger,
	)

	vad.Start()
	defer vad.Stop()

	// Create input stream
	inputStream, err := audio.NewInputStream(
		logger,
		audioOpts,
		func(buffer []float32) {
			vad.Feed(buffer)
		},
	)
	if err != nil {
		logger.Error("Failed to create input stream", "error", err)
		return
	}

	// Start input stream
	err = inputStream.Start()
	if err != nil {
		logger.Error("Failed to start input stream", "error", err)
		return
	}

	// Handle Ctrl+C to stop
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Stop and close input stream
	if err := inputStream.Stop(); err != nil {
		logger.Error("Failed to stop input stream", "error", err)
	}

	if err := inputStream.Close(); err != nil {
		logger.Error("Failed to close input stream", "error", err)
	}

	logger.Info("Program terminated gracefully")
}
