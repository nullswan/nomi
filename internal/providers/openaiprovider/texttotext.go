package openaiprovider

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/nullswan/golem/internal/chat"
	"github.com/nullswan/golem/internal/completion"
	baseprovider "github.com/nullswan/golem/internal/providers/base"
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
) (baseprovider.TextToTextProvider, error) {
	if config.model == "" {
		config.model = OpenAITextToTextDefaultModelFast
	}

	p := &TextToTextProvider{
		config: config,
		client: openai.NewClient(config.apiKey),
	}

	// Avoid checking model if using default model
	if config.model == OpenAITextToTextDefaultModelFast ||
		config.model == OpenAITextToTextDefaultModel {
		return p, nil
	}

	models, err := p.client.ListModels(context.Background())
	if err != nil {
		return nil, errors.New("error listing models")
	}

	for _, model := range models.Models {
		if model.ID == config.model {
			return p, nil
		}
	}

	return nil, fmt.Errorf("model %s not found", config.model)
}

func (p TextToTextProvider) Close() error {
	return nil
}

func (p TextToTextProvider) GetModel() string {
	return p.config.model
}

func (p TextToTextProvider) GenerateCompletion(
	ctx context.Context,
	messages []chat.Message,
	completionCh chan<- completion.Completion,
) error {
	req := completionRequestTextToText(p.config.model, messages)
	stream, err := p.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("error creating completion stream: %w", err)
	}

	aggCompletion := ""
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error receiving completion: %w", err)
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
