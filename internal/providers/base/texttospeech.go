package baseprovider

import "context"

type TextToSpeechProvider interface {
	GenerateSpeech(ctx context.Context, message string) ([]byte, error)

	GetModel() string
	Close() error
}
