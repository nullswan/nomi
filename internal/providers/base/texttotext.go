package baseprovider

import (
	"context"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/completion"
)

type TextToTextProvider interface {
	GenerateCompletion(
		ctx context.Context,
		messages []chat.Message,
		completionCh chan<- completion.Completion,
	) error

	GetModel() string
	Close() error
}
