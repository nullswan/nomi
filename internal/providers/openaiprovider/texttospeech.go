package openaiprovider

import (
	"context"
	"fmt"
	"io"

	baseprovider "github.com/nullswan/nomi/internal/providers/base"
	"github.com/sashabaranov/go-openai"
)

const (
	OpenAITextToSpeechDefaultModel = openai.TTSModel1
)

type TextToSpeechProvider struct {
	client *openai.Client
}

func NewTextToSpeechProvider(
	config oaiProviderConfig,
) (baseprovider.TextToSpeechProvider, error) {
	p := &TextToSpeechProvider{
		client: openai.NewClient(config.apiKey),
	}

	return p, nil
}

func (p TextToSpeechProvider) Close() error {
	return nil
}

func (p TextToSpeechProvider) GenerateSpeech(
	ctx context.Context,
	message string,
) ([]byte, error) {
	resp, err := p.client.CreateSpeech(ctx, openai.CreateSpeechRequest{
		Model: OpenAITextToSpeechDefaultModel,
		Voice: openai.VoiceAlloy,
		Input: message,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating speech: %w", err)
	}

	defer resp.Close()

	buf, err := io.ReadAll(resp)
	if err != nil {
		return nil, fmt.Errorf("error reading speech response: %w", err)
	}

	return buf, nil
}

// For now, we are always using the default model
func (p TextToSpeechProvider) GetModel() string {
	return string(OpenAITextToSpeechDefaultModel)
}
