package openrouterprovider

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/completion"
	baseprovider "github.com/nullswan/nomi/internal/providers/base"
	"github.com/sashabaranov/go-openai"
)

const (
	OpenAITextToJSONDefaultModel     = "openai/gpt-4o"
	OpenAITextToJSONDefaultModelFast = "openai/gpt-4o-mini"
)

type TextToJSONProvider struct {
	config openRouterProviderConfig
	client *openai.Client
}

func NewTextToJSONProvider(
	config openRouterProviderConfig,
) (baseprovider.TextToJSONProvider, error) {
	if config.model == "" {
		config.model = OpenAITextToJSONDefaultModelFast
	}

	oaConfig := openai.DefaultConfig(config.apiKey)
	oaConfig.BaseURL = baseURL

	oaClient := openai.NewClientWithConfig(oaConfig)

	p := &TextToJSONProvider{
		config: config,
		client: oaClient,
	}

	// Avoid checking model if using default model
	if config.model == OpenAITextToJSONDefaultModelFast ||
		config.model == OpenAITextToJSONDefaultModel {
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

func (p TextToJSONProvider) Close() error {
	return nil
}

func (p TextToJSONProvider) GetModel() string {
	return p.config.model
}

func (p TextToJSONProvider) GenerateCompletion(
	ctx context.Context,
	messages []chat.Message,
	completionCh chan<- completion.Completion,
) error {
	req := completionRequestTextToJSON(p.config.model, messages)
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

func completionRequestTextToJSON(
	model string,
	messages []chat.Message,
) openai.ChatCompletionRequest {
	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: make([]openai.ChatCompletionMessage, len(messages)),
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	}

	for i, message := range messages {
		req.Messages[i] = openai.ChatCompletionMessage{
			Role:    message.Role.String(),
			Content: message.Content,
		}
	}

	return req
}
