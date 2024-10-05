package openaiprovider

import (
	"context"

	"github.com/nullswan/golem/internal/chat"
	"github.com/nullswan/golem/internal/completion"
	provider "github.com/nullswan/golem/internal/providers/base"
	"github.com/sashabaranov/go-openai"
)

const (
	OpenAITextToTextReasoningDefaultModel     = openai.O1Preview
	OpenAITextToTextReasoningDefaultModelFast = openai.O1Mini
)

type TextToTextReasoningProvider struct{}

func NewTextToTextReasoningProvider(
	apiKey string,
) provider.TextToTextReasoningProvider {
	return &TextToTextReasoningProvider{}
}

func (p *TextToTextReasoningProvider) GenerateCompletion(
	ctx context.Context,
	messages []chat.Message,
	completionCh chan<- completion.Completion,
) error {
	return nil
}
