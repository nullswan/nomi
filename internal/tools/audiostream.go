package tools

import (
	"fmt"
	"log/slog"

	"github.com/gordonklaus/portaudio"
	"github.com/nullswan/nomi/internal/audio"
)

type AudioStream interface {
	Start() error
	Close() error
}

type audioStream struct {
	stream *audio.StreamHandler
}

func (a *audioStream) Close() error {
	err := a.stream.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop audio stream: %w", err)
	}
	return nil
}

func (a *audioStream) Start() error {
	err := a.stream.Start()
	if err != nil {
		return fmt.Errorf("failed to start audio stream: %w", err)
	}
	return nil
}

func NewAudioStream(
	logger *slog.Logger,
	device *portaudio.DeviceInfo,
	callback func([]float32), // TODO(nullswan): Make registrable instead of passed
) (AudioStream, error) {
	stream, err := audio.NewInputStream(
		logger,
		device,
		&audio.StreamParameters{},
		callback,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio stream: %w", err)
	}

	return &audioStream{
		stream: stream,
	}, nil
}
