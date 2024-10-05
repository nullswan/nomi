package olamalocalprovider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/nullswan/golem/internal/chat"
	"github.com/nullswan/golem/internal/completion"
	provider "github.com/nullswan/golem/internal/providers/base"
	"github.com/ollama/ollama/api"
)

const (
	OLamaTextToTextDefaultModel     = "llama3.1"
	OLamaTextToTextDefaultModelFast = "llama3.2"
)

type TextToTextProvider struct {
	config olamaProviderConfig
	client *api.Client
}

func NewTextToTextProvider(
	config olamaProviderConfig,
) provider.TextToTextProvider {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	url, err := url.Parse(config.BaseUrl())
	if err != nil {
		panic(err)
	}

	if config.model == "" {
		config.model = OLamaTextToTextDefaultModelFast
	}

	return &TextToTextProvider{
		config: config,
		client: api.NewClient(
			url,
			httpClient,
		),
	}
}

func (p *TextToTextProvider) GenerateCompletion(
	ctx context.Context,
	messages []chat.Message,
	completionCh chan<- completion.Completion,
) error {
	req := completionRequestTextToText(p.config.model, messages)

	aggCompletion := ""
	resp := func(resp api.ChatResponse) error {
		if resp.Done {
			completionCh <- completion.NewCompletionTombStone(
				aggCompletion,
				p.config.model,
				completion.Usage{},
			)
			return nil
		}

		completionCh <- completion.NewCompletionData(
			resp.Message.Content,
		)

		return nil
	}

	err := p.client.Chat(ctx, &req, resp)
	if err != nil {
		return fmt.Errorf("error creating completion stream: %w", err)
	}

	return nil
}

func completionRequestTextToText(
	model string,
	messages []chat.Message,
) api.ChatRequest {
	stream := true

	req := api.ChatRequest{
		Model:    model,
		Stream:   &stream,
		Messages: make([]api.Message, len(messages)),
	}

	for i, m := range messages {
		req.Messages[i] = api.Message{
			Content: m.Content,
			Role:    m.Role.String(),
		}
	}

	return req
}
