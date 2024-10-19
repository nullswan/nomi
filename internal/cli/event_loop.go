package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/glamour"
	"github.com/chzyer/readline"
	"github.com/nullswan/nomi/internal/audio"
	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/logger"
	baseprovider "github.com/nullswan/nomi/internal/providers/base"
	"github.com/nullswan/nomi/internal/term"
)

// TODO(nullswan): Refactor this to use a more generic function signature.
type ProcessInputFuncT func(context.Context, string, chat.Conversation, *glamour.TermRenderer, baseprovider.TextToTextProvider, *readline.Instance)

// EventLoop manages the main event loop.
func EventLoop(
	ctx context.Context,
	cancel context.CancelFunc,
	inputCh chan string,
	inputErrCh chan error,
	audioStartCh, audioEndCh <-chan struct{},
	inputStream *audio.AudioStream,
	log *logger.Logger,
	conversation chat.Conversation,
	renderer *term.Renderer,
	textToTextBackend baseprovider.TextToTextProvider,
	rl *readline.Instance,
	processInputFunc ProcessInputFuncT,
) {
	audioRunning := false

	defer func() {
		if audioRunning {
			inputStream.Stop()
		}
	}()

	eventCtx, eventCtxCancel := context.WithCancel(ctx)
	defer eventCtxCancel()

	for {
		select {
		case <-ctx.Done():
			return
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
				rl,
			)
		case err := <-inputErrCh:
			if errors.Is(err, term.ErrInputInterrupted) ||
				errors.Is(err, term.ErrInputKilled) {
				cancel()
				return
			}
			fmt.Printf("Error reading input: %v\n", err)
		case <-audioStartCh:
			// TODO(nullswan): Add graphical feedback for audio recording.
			if !audioRunning {
				audioRunning = true
				err := inputStream.Start()
				if err != nil {
					log.With("error", err).Error("Failed to start input stream")
				}
			}
		case <-audioEndCh:
			// TODO(nullswan): Add graphical feedback for audio recording.
			if audioRunning {
				audioRunning = false
				err := inputStream.Stop()
				if err != nil {
					log.With("error", err).Error("Failed to stop input stream")
				}
			}
		}
	}
}
