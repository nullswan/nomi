package prompts

import (
	"os"
	"path/filepath"

	"github.com/nullswan/ai/internal/config"
	"gopkg.in/yaml.v2"
)

func LoadPrompt(filename string) (*Prompt, error) {
	dataDir := config.GetDataDir()
	data, err := os.ReadFile(filepath.Join(dataDir, filename))
	if err != nil {
		return nil, err
	}

	var prompt Prompt
	err = yaml.Unmarshal(data, &prompt)
	if err != nil {
		return nil, err
	}
	return &prompt, nil
}

func ListPrompts() ([]Prompt, error) {
	dataDir := config.GetDataDir()
	files, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	prompts := DefaultPrompts
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !IsPromptFile(file.Name()) {
			continue
		}

		prompt, err := LoadPrompt(file.Name())
		if err != nil {
			return nil, err
		}
		prompts = append(prompts, *prompt)
	}

	return prompts, nil
}

func IsPromptFile(filename string) bool {
	return false
}
