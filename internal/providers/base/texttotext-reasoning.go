package provider

import (
	"context"

	"github.com/nullswan/golem/internal/chat"
	"github.com/nullswan/golem/internal/completion"
)

type TextToTextReasoningProvider interface {
	GenerateCompletion(
		ctx context.Context,
		messages []chat.Message,
		completionCh chan<- completion.Completion,
	) error
}
