package openaiprovider

import (
	"context"

	"github.com/nullswan/golem/internal/chat"
	"github.com/nullswan/golem/internal/completion"
	provider "github.com/nullswan/golem/internal/providers/base"
	"github.com/sashabaranov/go-openai"
)

const (
	OpenAITextToTextDefaultModel     = openai.GPT4o
	OpenAITextToTextDefaultModelFast = openai.GPT4oMini
)

type TextToTextProvider struct{}

func NewTextToTextProvider(
	apiKey string,
) provider.TextToTextProvider {
	return &TextToTextProvider{}
}

func (p *TextToTextProvider) GenerateCompletion(
	ctx context.Context,
	messages []chat.Message,
	completionCh chan<- completion.Completion,
) error {
	return nil
}
