package tools

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/nullswan/nomi/internal/audio"
	"github.com/nullswan/nomi/internal/term"
)

// TODO(nullswan): Handle Command handling
type InputHandler interface {
	Read(ctx context.Context, defaultValue string) (string, error)

	WithVoiceInput(
		voiceInputCh chan string,
		audioStartCh <-chan struct{},
		audioEndCh <-chan struct{},
		inputStream *audio.StreamHandler,
	)
}

type inputHandler struct {
	logger *slog.Logger

	voiceInputCh chan string
	audioStartCh <-chan struct{}
	audioEndCh   <-chan struct{}

	inputStream *audio.StreamHandler
}

func NewInputHandler(
	logger *slog.Logger,
) InputHandler {
	return &inputHandler{
		logger:       logger,
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
	inputStream *audio.StreamHandler,
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
	rl, err := term.InitReadline(defaultValue)
	if err != nil {
		return "", fmt.Errorf("error initializing readline: %w", err)
	}

	inputErrCh := make(chan error)
	inputCh := make(chan string)

	go func() {
		ret, err := term.ReadInputOnce(rl)
		if err != nil {
			if rl.Closed() {
				return
			}

			select {
			case inputErrCh <- err:
			case <-ctx.Done():
			}

			return
		}

		select {
		case inputCh <- ret:
		case <-ctx.Done():
			return
		}
	}()

	audioRunning := false
	spinner := term.NewSpinner(1*time.Second, ">>> ")

	defer func() {
		if audioRunning {
			i.inputStream.Stop() // nolint:errcheck
		}
		rl.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return "", errors.New("context canceled")
		case line := <-i.voiceInputCh:
			return line, nil
		case line := <-inputCh:
			return line, nil
		case err := <-inputErrCh:
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
