package tools

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nullswan/nomi/internal/audio"
	"github.com/nullswan/nomi/internal/term"
)

type InputHandler interface {
	Read(ctx context.Context, defaultValue string) (string, error)

	WithVoiceInput(
		voiceInputCh chan string,
		audioStartCh <-chan struct{},
		audioEndCh <-chan struct{},
		inputStream *audio.AudioStream,
	)
}

type inputHandler struct {
	logger *slog.Logger

	readyCh    chan struct{}
	inputCh    chan string
	inputErrCh chan error

	voiceInputCh chan string
	audioStartCh <-chan struct{}
	audioEndCh   <-chan struct{}

	inputStream *audio.AudioStream
}

func NewInputHandler(
	logger *slog.Logger,
	readyCh chan struct{},
	inputCh chan string,
	inputErrCh chan error,
) InputHandler {
	return &inputHandler{
		logger:       logger,
		readyCh:      readyCh,
		inputCh:      inputCh,
		inputErrCh:   inputErrCh,
		voiceInputCh: nil,
		audioStartCh: nil,
		audioEndCh:   nil,
		inputStream:  nil,
	}
}

func (i *inputHandler) WithVoiceInput(
	voiceInputCh chan string,
	audioStartCh <-chan struct{},
	audioEndCh <-chan struct{},
	inputStream *audio.AudioStream,
) {
	i.voiceInputCh = voiceInputCh
	i.audioStartCh = audioStartCh
	i.audioEndCh = audioEndCh
	i.inputStream = inputStream
}

func (i *inputHandler) Read(
	ctx context.Context,
	defaultValue string,
) (string, error) {
	select {
	case i.readyCh <- struct{}{}:
	default:
	}

	audioRunning := false
	spinner := term.NewSpinner(1*time.Second, ">>> ")

	defer func() {
		if audioRunning {
			i.inputStream.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("context canceled")
		case line := <-i.voiceInputCh:
			return line, nil
		case line := <-i.inputCh:
			return line, nil
		case err := <-i.inputErrCh:
			return "", fmt.Errorf("error reading input: %w", err)
		case <-i.audioStartCh:
			if !audioRunning {
				audioRunning = true
				err := i.inputStream.Start()
				if err != nil {
					i.logger.With("error", err).
						Error("Failed to start input stream")
				} else {
					spinner.Start()
				}
			}
		case <-i.audioEndCh:
			if audioRunning {
				audioRunning = false
				err := i.inputStream.Stop()
				if err != nil {
					i.logger.With("error", err).
						Error("Failed to stop input stream")
				} else {
					spinner.Stop()
				}
			}
		}
	}
}
