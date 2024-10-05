package openaiprovider

import (
	"context"
	"io"

	"github.com/nullswan/golem/internal/chat"
	"github.com/nullswan/golem/internal/completion"
	provider "github.com/nullswan/golem/internal/providers/base"
	"github.com/sashabaranov/go-openai"
)

const (
	OpenAITextToTextDefaultModel     = openai.GPT4o
	OpenAITextToTextDefaultModelFast = openai.GPT4oMini
)

type TextToTextProvider struct {
	config oaiProviderConfig
	client *openai.Client
}

func NewTextToTextProvider(
	config oaiProviderConfig,
) provider.TextToTextProvider {
	if config.model == "" {
		config.model = OpenAITextToTextDefaultModelFast
	}

	return &TextToTextProvider{
		config: config,
		client: openai.NewClient(config.apiKey),
	}
}

func (p *TextToTextProvider) GenerateCompletion(
	ctx context.Context,
	messages []chat.Message,
	completionCh chan<- completion.Completion,
) error {
	req := completionRequestTextToText(p.config.model, messages)
	stream, err := p.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return err
	}

	aggCompletion := ""
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		completionCh <- completion.NewCompletionData(
			resp.Choices[0].Delta.Content,
		)

		aggCompletion += resp.Choices[0].Delta.Content
	}

	completionCh <- completion.NewCompletionTombStone(
		aggCompletion,
		p.config.model,
		completion.Usage{},
	)

	return nil
}

func completionRequestTextToText(
	model string,
	messages []chat.Message,
) openai.ChatCompletionRequest {
	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: make([]openai.ChatCompletionMessage, len(messages)),
	}

	for i, message := range messages {
		req.Messages[i] = openai.ChatCompletionMessage{
			Role:    message.Role.String(),
			Content: message.Content,
		}
	}

	return req
}