package baseprovider

import (
	"context"

	"github.com/nullswan/golem/internal/chat"
	"github.com/nullswan/golem/internal/completion"
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
