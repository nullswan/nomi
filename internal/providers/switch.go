package providers

import "os"

func FindFirstProvider() string {
	provider := "ollama"
	if os.Getenv("OPENAI_API_KEY") != "" {
		provider = "openai"
	}

	return provider
}
