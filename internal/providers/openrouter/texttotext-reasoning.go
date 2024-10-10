package openrouterprovider

import (
	"context"
	"fmt"

	"github.com/nullswan/golem/internal/chat"
	"github.com/nullswan/golem/internal/completion"
	baseprovider "github.com/nullswan/golem/internal/providers/base"
	"github.com/sashabaranov/go-openai"
)

const (
	OpenAITextToTextReasoningDefaultModel     = "openai/o1-preview"
	OpenAITextToTextReasoningDefaultModelFast = "openai/o1-mini"
)

type TextToTextReasoningProvider struct {
	config openRouterProviderConfig
	client *openai.Client
}

func NewTextToTextReasoningProvider(
	config openRouterProviderConfig,
) (baseprovider.TextToTextReasoningProvider, error) {
	if config.model == "" {
		config.model = OpenAITextToTextReasoningDefaultModelFast
	}

	oaConfig := openai.DefaultConfig(config.apiKey)
	oaConfig.BaseURL = baseURL

	oaClient := openai.NewClientWithConfig(oaConfig)

	p := &TextToTextReasoningProvider{
		config: config,
		client: oaClient,
	}

	if config.model == OpenAITextToTextReasoningDefaultModelFast ||
		config.model == OpenAITextToTextReasoningDefaultModel {
		return p, nil
	}

	models, err := p.client.ListModels(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error listing models: %w", err)
	}

	for _, model := range models.Models {
		if model.ID == config.model {
			return p, nil
		}
	}

	return nil, fmt.Errorf("model %s not found", config.model)
}

func (p *TextToTextReasoningProvider) Close() error {
	return nil
}

func (p *TextToTextReasoningProvider) GetModel() string {
	return p.config.model
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
		return fmt.Errorf("error creating completion stream: %w", err)
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

	for _, message := range messages {
		role := message.Role.String()
		// System prompt are not supported YET (cf: https://platform.openai.com/docs/guides/reasoning/beta-limitations)
		if message.Role == chat.RoleSystem {
			role = chat.RoleUser.String()
		}

		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: message.Content,
		})
	}

	return req
}
