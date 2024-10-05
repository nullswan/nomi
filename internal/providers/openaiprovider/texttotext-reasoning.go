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

type TextToTextReasoningProvider struct {
	config oaiProviderConfig
	client *openai.Client
}

func NewTextToTextReasoningProvider(
	config oaiProviderConfig,
) provider.TextToTextReasoningProvider {
	if config.model == "" {
		config.model = OpenAITextToTextReasoningDefaultModel
	}

	return &TextToTextReasoningProvider{
		config: config,
		client: openai.NewClient(config.apiKey),
	}
}

func (p *TextToTextReasoningProvider) GenerateCompletion(
	ctx context.Context,
	messages []chat.Message,
	completionCh chan<- completion.Completion,
) error {
	req := completionRequestTextToTextReasoning(p.config.model, messages)

	// Streaming is not supported YET (cf: https://platform.openai.com/docs/guides/reasoning/beta-limitations)
	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return err
	}

	completionCh <- completion.NewCompletionData(
		resp.Choices[0].Message.Content,
	)

	completionCh <- completion.NewCompletionTombStone(
		resp.Choices[0].Message.Content,
		p.config.model,
		completion.Usage{},
	)

	return nil
}

func completionRequestTextToTextReasoning(
	model string,
	messages []chat.Message,
) openai.ChatCompletionRequest {
	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: []openai.ChatCompletionMessage{},
	}

	// System prompt are not supported YET (cf: https://platform.openai.com/docs/guides/reasoning/beta-limitations)
	for _, message := range messages {
		if message.Role != chat.RoleUser {
			// TODO(nullswan): Add logger
			continue
		}

		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    message.Role.String(),
			Content: message.Content,
		})
	}

	return req
}
