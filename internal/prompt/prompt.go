package prompts

import "time"

type Prompt struct {
	ID          string      `yaml:"id"`
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Settings    Settings    `yaml:"settings"`
	Metadata    Metadata    `yaml:"metadata"`
	Preferences Preferences `yaml:"preferences"`
}

type Settings struct {
	SystemPrompt string  `yaml:"system_prompt"`
	PrePrompt    *string `yaml:"pre_prompt"`
}

type Preferences struct {
	Fast      bool `yaml:"fast"`
	Reasoning bool `yaml:"reasoning"`
}

type Metadata struct {
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
	Version   string    `yaml:"version"`
	Author    string    `yaml:"author"`
}
