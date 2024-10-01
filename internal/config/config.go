package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var configFilePath string

func init() {
	configFilePath = getConfigFilePath()
}

func getConfigFilePath() string {
	return filepath.Join(GetDataDir(), "config.yml")
}

// ConfigExists checks if the configuration file exists.
func ConfigExists() bool {
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// LoadConfig loads the configuration from the YAML file or creates a default one if it doesn't exist.
func LoadConfig() (*Config, error) {
	var cfg Config

	if !ConfigExists() {
		// File does not exist, create default configuration
		cfg = defaultConfig()
		if err := SaveConfig(&cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveConfig saves the configuration to the YAML file.
func SaveConfig(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(configFilePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

const (
	defaultConversationDir = "conversations"
)

// defaultConfig returns the default configuration.
func defaultConfig() Config {
	convDir := GetDataSubdir(defaultConversationDir)

	return Config{
		Input: InputConfig{
			Text:  EnabledConfig{Enabled: true},
			Image: EnabledConfig{Enabled: true},
			Voice: EnabledConfig{Enabled: true},
		},
		Output: OutputConfig{
			Markdown: OutputDetailConfig{
				Enabled: true,
				Path:    filepath.Join(convDir, "markdown"),
			},
			Sqlite: OutputDetailConfig{
				Enabled: true,
				Path:    filepath.Join(convDir, "output.sqlite.db"),
			},
		},
		Memory: MemoryConfig{
			Mode: MemoryModeNone,
		},
	}
}
