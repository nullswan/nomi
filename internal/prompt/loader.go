package prompts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nullswan/nomi/internal/config"
	"gopkg.in/yaml.v2"
)

var ErrPromptNotFound = errors.New("prompt not found")

func LoadPrompt(filename string) (*Prompt, error) {
	if !strings.HasSuffix(filename, ".yml") {
		filename += ".yml"
	}

	fp := filepath.Join(config.GetPromptDirectory(), filename)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return nil, ErrPromptNotFound
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		return nil, fmt.Errorf("error reading prompt file: %w", err)
	}

	var prompt Prompt
	err = yaml.Unmarshal(data, &prompt)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling prompt file: %w", err)
	}

	return &prompt, nil
}

func ListPrompts() ([]Prompt, error) {
	files, err := os.ReadDir(config.GetPromptDirectory())
	if err != nil {
		return nil, fmt.Errorf("error reading data directory: %w", err)
	}

	var prompts []Prompt
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Prompt files are YAML files
		if !isValidFilename(file.Name()) {
			continue
		}

		prompt, err := LoadPrompt(file.Name())
		if err != nil {
			return nil, fmt.Errorf("error loading prompt: %w", err)
		}

		if err := prompt.Validate(); err != nil {
			return nil, fmt.Errorf("error validating prompt: %w", err)
		}

		prompts = append(prompts, *prompt)
	}

	return prompts, nil
}

func isValidFilename(filename string) bool {
	return filepath.Ext(filename) == ".yml"
}
