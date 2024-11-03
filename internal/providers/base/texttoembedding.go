package baseprovider

import "context"

type TextToEmbeddingProvider interface {
	GenerateEmbedding(
		ctx context.Context,
		message string,
	) ([]float32, error)

	GetModel() string
	Close() error
}
