package cli

import (
	"fmt"

	"github.com/gordonklaus/portaudio"
	"github.com/nullswan/nomi/internal/audio"
	"github.com/nullswan/nomi/internal/logger"
)

// InitPortAudio initializes the PortAudio library.
func InitPortAudio() error {
	if err := portaudio.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize PortAudio: %w", err)
	}
	return nil
}

// TerminatePortAudio terminates the PortAudio library.
func TerminatePortAudio() error {
	if err := portaudio.Terminate(); err != nil {
		return fmt.Errorf("failed to terminate PortAudio: %w", err)
	}
	return nil
}

// ComputeAudioOptions computes and returns the audio options.
func ComputeAudioOptions(
	opts *audio.AudioOptions,
) (*audio.AudioOptions, error) {
	computedOpts, err := audio.ComputeAudioOptions(opts)
	if err != nil {
		return nil, fmt.Errorf("error computing audio options: %w", err)
	}
	return computedOpts, nil
}

// InitializeAudioStream initializes and returns the audio input stream.
func InitializeAudioStream(
	log *logger.Logger,
	audioOpts *audio.AudioOptions,
	vad *audio.VAD,
) (*audio.AudioStream, error) {
	inputStream, err := audio.NewInputStream(
		log,
		audioOpts,
		func(buffer []float32) {
			vad.Feed(buffer)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio stream: %w", err)
	}
	return inputStream, nil
}
