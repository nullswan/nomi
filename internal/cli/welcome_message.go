package cli

import (
	"fmt"

	"github.com/nullswan/nomi/internal/providers"
)

type ModelProvider interface {
	GetModel() string
}

type Conversation interface {
	GetID() string
}

type WelcomeConfig struct {
	Conversation    Conversation
	Provider        []providers.AIProvider
	ModelProviders  []ModelProvider
	WelcomeMessage  string
	StartPrompt     *string
	BuildVersion    string
	BuildDate       string
	AdditionalLines []string
	Instructions    []string
}

type WelcomeOption func(*WelcomeConfig)

func WithModelProvider(mp ModelProvider) WelcomeOption {
	return func(c *WelcomeConfig) {
		c.ModelProviders = append(c.ModelProviders, mp)
	}
}

func WithProvider(provider providers.AIProvider) WelcomeOption {
	return func(c *WelcomeConfig) {
		c.Provider = append(c.Provider, provider)
	}
}

func WithWelcomeMessage(msg string) WelcomeOption {
	return func(c *WelcomeConfig) {
		c.WelcomeMessage = msg
	}
}

func WithStartPrompt(prompt string) WelcomeOption {
	return func(c *WelcomeConfig) {
		c.StartPrompt = &prompt
	}
}

func WithBuildVersion(version string) WelcomeOption {
	return func(c *WelcomeConfig) {
		c.BuildVersion = version
	}
}

func WithBuildDate(date string) WelcomeOption {
	return func(c *WelcomeConfig) {
		c.BuildDate = date
	}
}

func WithAdditionalLine(line string) WelcomeOption {
	return func(c *WelcomeConfig) {
		c.AdditionalLines = append(c.AdditionalLines, line)
	}
}

func WithInstruction(instr string) WelcomeOption {
	return func(c *WelcomeConfig) {
		c.Instructions = append(c.Instructions, instr)
	}
}

func NewWelcomeConfig(
	conversation Conversation,
	opts ...WelcomeOption,
) WelcomeConfig {
	config := WelcomeConfig{
		Conversation:    conversation,
		Provider:        []providers.AIProvider{},
		ModelProviders:  []ModelProvider{},
		AdditionalLines: []string{},
		Instructions:    []string{},
	}
	for _, opt := range opts {
		opt(&config)
	}
	return config
}

func DisplayWelcome(config WelcomeConfig) {
	fmt.Printf("----\n")
	fmt.Println(config.WelcomeMessage)
	fmt.Println()
	fmt.Println("Configuration")
	if config.StartPrompt != nil {
		fmt.Printf("  Start prompt: %s\n", *config.StartPrompt)
	}
	for i, mp := range config.ModelProviders {
		fmt.Printf("  Model %d: %s\n", i+1, mp.GetModel())
	}
	fmt.Printf("  Conversation: %s\n", config.Conversation.GetID())
	for i, provider := range config.Provider {
		fmt.Printf("  Provider: %d: %s\n", i+1, provider.String())
	}
	fmt.Printf("  Build Version: %s\n", config.BuildVersion)
	fmt.Printf("  Build Date: %s\n", config.BuildDate)
	for _, line := range config.AdditionalLines {
		fmt.Println(line)
	}
	fmt.Printf("-----\n")
	for _, instr := range config.Instructions {
		fmt.Println(instr)
	}
	fmt.Printf("-----\n\n")
}
