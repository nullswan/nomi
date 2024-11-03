package openaiprovider

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

const (
	OpenAITextToEmbeddingDefaultModel = openai.LargeEmbedding3
	OpenAITextToEmbeddingFastModel    = openai.SmallEmbedding3
)

type TextToEmbeddingProvider struct {
	client *openai.Client
}

func NewTextToEmbeddingProvider(
	config oaiProviderConfig,
) (TextToEmbeddingProvider, error) {
	p := TextToEmbeddingProvider{
		client: openai.NewClient(config.apiKey),
	}

	return p, nil
}

func (p TextToEmbeddingProvider) Close() error {
	return nil
}

func (p TextToEmbeddingProvider) GenerateEmbedding(
	ctx context.Context,
	message string,
) ([]float32, error) {
	resp, err := p.client.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequest{
			Model: OpenAITextToEmbeddingDefaultModel,
			Input: message,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error creating embedding: %w", err)
	}

	return resp.Data[0].Embedding, nil
}

// For now, we are always using the default model
func (p TextToEmbeddingProvider) GetModel() string {
	return string(OpenAITextToEmbeddingDefaultModel)
}
