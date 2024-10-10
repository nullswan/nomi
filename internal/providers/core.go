package providers

import "os"

type AIProvider string

const (
	OpenAIProvider     AIProvider = "openai"
	AnthropicProvider  AIProvider = "anthropic"
	OpenRouterProvider AIProvider = "openrouter"
	OllamaProvider     AIProvider = "ollama"
)

func (p AIProvider) String() string {
	return string(p)
}

func CheckProvider() AIProvider {
	if os.Getenv("OPENAI_API_KEY") != "" {
		return OpenAIProvider
	}

	if os.Getenv("OPENROUTER_API_KEY") != "" {
		return OpenRouterProvider
	}

	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return AnthropicProvider
	}

	return OllamaProvider
}
