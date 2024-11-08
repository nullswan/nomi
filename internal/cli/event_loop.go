package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/nullswan/nomi/internal/audio"
	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/logger"
	baseprovider "github.com/nullswan/nomi/internal/providers/base"
	"github.com/nullswan/nomi/internal/term"
)

// TODO(nullswan): Refactor this to use a more generic function signature.
type ProcessInputFuncT func(context.Context, string, chat.Conversation, *glamour.TermRenderer, baseprovider.TextToTextProvider)

// EventLoop manages the main event loop.
func EventLoop(
	ctx context.Context,
	cancel context.CancelFunc,
	inputCh chan string,
	inputErrCh chan error,
	readyCh chan struct{},
	voiceInputCh chan string,
	audioStartCh, audioEndCh <-chan struct{},
	inputStream *audio.StreamHandler,
	log *logger.Logger,
	conversation chat.Conversation,
	renderer *term.Renderer,
	textToTextBackend baseprovider.TextToTextProvider,
	processInputFunc ProcessInputFuncT,
) {
	audioRunning := false
	spinner := term.NewSpinner(1*time.Second, ">>> ")

	defer func() {
		if audioRunning {
			inputStream.Stop()
		}
	}()

	eventCtx, eventCtxCancel := context.WithCancel(ctx)
	defer eventCtxCancel()

	// Signal that the event loop is ready to receive input.
	readyCh <- struct{}{}

	for {
		select {
		case <-ctx.Done():
			return
		case line := <-voiceInputCh:
			eventCtxCancel()
			eventCtx, eventCtxCancel = context.WithCancel(ctx)
			defer eventCtxCancel()

			processInputFunc(
				eventCtx,
				line,
				conversation,
				renderer,
				textToTextBackend,
			)

			// Signal that the event loop is ready to receive input again.
			select {
			case readyCh <- struct{}{}:
			default:
				fmt.Printf(">>> ")
				continue
			}
		case line := <-inputCh:
			eventCtxCancel()
			eventCtx, eventCtxCancel = context.WithCancel(ctx)
			defer eventCtxCancel()

			processInputFunc(
				eventCtx,
				line,
				conversation,
				renderer,
				textToTextBackend,
			)

			// Signal that the event loop is ready to receive input again.
			select {
			case readyCh <- struct{}{}:
			default:
				continue
			}

		case err := <-inputErrCh:
			if errors.Is(err, term.ErrInputInterrupted) ||
				errors.Is(
					err,
					term.ErrInputKilled,
				) || errors.Is(err, term.ErrReadlineInit) {
				cancel()
				return
			}
			fmt.Printf("Error reading input: %v\n", err)
		case <-audioStartCh:
			if !audioRunning {
				audioRunning = true
				err := inputStream.Start()
				if err != nil {
					log.With("error", err).Error("Failed to start input stream")
				} else {
					spinner.Start()
				}
			}
		case <-audioEndCh:
			if audioRunning {
				audioRunning = false
				err := inputStream.Stop()
				if err != nil {
					log.With("error", err).Error("Failed to stop input stream")
				} else {
					spinner.Stop()
				}
			}
		}
	}
}
