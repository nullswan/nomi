package prompts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nullswan/golem/internal/config"
	"gopkg.in/yaml.v2"
)

func (p *Prompt) Validate() error {
	if p.ID == "" {
		return errors.New("Prompt ID is required")
	}

	if p.Name == "" {
		return errors.New("Prompt Name is required")
	}

	if p.Settings.SystemPrompt == "" {
		return errors.New("Prompt SystemPrompt is required")
	}

	if p.Metadata.Version == "" {
		return errors.New("Prompt Version is required")
	}

	if p.Metadata.Author == "" {
		return errors.New("Prompt Author is required")
	}

	if p.Metadata.CreatedAt.IsZero() {
		return errors.New("Prompt CreatedAt is required")
	}

	if p.Metadata.UpdatedAt.IsZero() {
		return errors.New("Prompt UpdatedAt is required")
	}

	return nil
}

// Save a prompt to the disk
// TODO(nullswan): Use a store ID instead of the prompt ID
func (p *Prompt) Save() error {
	promptFile := filepath.Join(config.GetPromptDirectory(), p.ID+".yml")
	outData, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("Error marshalling prompt to YAML: %v", err)
	}

	err = os.WriteFile(promptFile, outData, 0o644)
	if err != nil {
		return fmt.Errorf("Error writing prompt file: %v", err)
	}

	return nil
}
