package tools

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/completion"
	baseprovider "github.com/nullswan/nomi/internal/providers/base"
)

type TextToTextBackend struct {
	backend baseprovider.TextToTextProvider
	logger  *slog.Logger
}

func NewTextToTextBackend(
	backend baseprovider.TextToTextProvider,
	logger *slog.Logger,
) TextToTextBackend {
	return TextToTextBackend{
		backend: backend,
		logger:  logger,
	}
}

func (t TextToTextBackend) Do(
	ctx context.Context,
	conversation chat.Conversation,
) (string, error) {
	messages := conversation.GetMessages()

	outCh := make(chan completion.Completion)
	go func() {
		defer close(outCh)
		if err := t.backend.GenerateCompletion(ctx, messages, outCh); err != nil {
			if strings.Contains(err.Error(), "context canceled") {
				return
			}
			t.logger.With("error", err).
				Error("Error generating completion")
		}
	}()

	for {
		select {
		case cmpl, ok := <-outCh:
			if !ok {
				return "", fmt.Errorf("completion channel closed")
			}
			if !completion.IsTombStone(cmpl) {
				continue
			}

			return cmpl.Content(), nil
		}
	}
}

type TextToJSONBackend struct {
	backend baseprovider.TextToJSONProvider
	logger  *slog.Logger
}

func NewTextToJSONBackend(
	backend baseprovider.TextToJSONProvider,
	logger *slog.Logger,
) TextToJSONBackend {
	return TextToJSONBackend{
		backend: backend,
		logger:  logger,
	}
}

func (t TextToJSONBackend) Do(
	ctx context.Context,
	conversation chat.Conversation,
) (string, error) {
	messages := conversation.GetMessages()

	outCh := make(chan completion.Completion)
	go func() {
		defer close(outCh)
		if err := t.backend.GenerateCompletion(ctx, messages, outCh); err != nil {
			if strings.Contains(err.Error(), "context canceled") {
				return
			}
			t.logger.With("error", err).
				Error("Error generating completion")
		}
	}()

	for {
		select {
		case cmpl, ok := <-outCh:
			if !ok {
				return "", fmt.Errorf("completion channel closed")
			}
			if !completion.IsTombStone(cmpl) {
				continue
			}

			content := strings.Replace(cmpl.Content(), "```json", "", -1)
			content = strings.Replace(content, "```", "", -1)
			return content, nil
		}
	}
}
