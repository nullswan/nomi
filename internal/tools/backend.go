package tools

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/completion"
	baseprovider "github.com/nullswan/nomi/internal/providers/base"
	"github.com/nullswan/nomi/internal/sound"
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

	for cmpl := range outCh {
		if !completion.IsTombStone(cmpl) {
			continue
		}
		return cmpl.Content(), nil
	}

	return "", errors.New("completion channel closed")
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

	for cmpl := range outCh {
		if !completion.IsTombStone(cmpl) {
			continue
		}

		content := strings.ReplaceAll(cmpl.Content(), "```json", "")
		return strings.ReplaceAll(content, "```", ""), nil
	}

	return "", errors.New("completion channel closed")
}

type TextToSpeechBackend struct {
	backend baseprovider.TextToSpeechProvider
	logger  *slog.Logger
}

func NewTextToSpeechBackend(
	backend baseprovider.TextToSpeechProvider,
	logger *slog.Logger,
) *TextToSpeechBackend {
	return &TextToSpeechBackend{
		backend: backend,
		logger:  logger,
	}
}

func (t TextToSpeechBackend) Do(
	ctx context.Context,
	message string,
) ([]byte, error) {
	buf, err := t.backend.GenerateSpeech(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("error generating speech: %w", err)
	}

	return buf, nil
}

func (t TextToSpeechBackend) Speak(
	ctx context.Context,
	message string,
) error {
	buf, err := t.Do(ctx, message)
	if err != nil {
		return err
	}

	err = sound.PlayBuffer(buf)
	if err != nil {
		return fmt.Errorf("error playing sound: %w",
			err)
	}

	return nil
}
