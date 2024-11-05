package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func getConfigFilePath() string {
	return filepath.Join(GetProgramDirectory(), configFileName)
}

// Exists checks if the configuration file exists.
func Exists() bool {
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// LoadConfig loads the configuration from the YAML file or creates a default one if it doesn't exist.
func LoadConfig() (*Config, error) {
	var cfg Config

	if !Exists() {
		// File does not exist, create default configuration
		cfg = DefaultConfig()
		if err := SaveConfig(&cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %w", err)
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf(
			"error unmarshalling configuration file: %w",
			err,
		)
	}

	return &cfg, nil
}

// SaveConfig saves the configuration to the YAML file.
func SaveConfig(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshalling configuration: %w", err)
	}

	err = os.WriteFile(configFilePath, data, 0o644)
	if err != nil {
		return fmt.Errorf("error writing configuration file: %w", err)
	}

	return nil
}

// defaultConfig returns the default configuration.
func DefaultConfig() Config {
	convDir := GetConversationDirectory()

	// Do nothing, just create the directory if it doesn't exist
	_ = GetPromptDirectory()

	const optionCmdDefault = 58

	return Config{
		DevMode: false,
		Input: InputConfig{
			Voice: VoiceConfig{
				Enabled:  false,
				Language: "en",
				KeyCode:  optionCmdDefault,
			},
		},
		Output: OutputConfig{
			Sqlite: SqliteConfig{
				Enabled: true,
				Path:    filepath.Join(convDir, "sqlite.db"),
			},
			Speech: SpeechConfig{
				Enabled: false,
			},
		},
		PlaySound: false,
	}
}
