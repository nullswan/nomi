package config

type Config struct {
	Input     InputConfig  `yaml:"input"      json:"input"`
	Output    OutputConfig `yaml:"output"     json:"output"`
	DevMode   bool         `yaml:"dev_mode"   json:"dev_mode"`
	PlaySound bool         `yaml:"play_sound" json:"play_sound"`
	// TODO(nullswan): Add memory configuration
}

// Manage the input sources
type InputConfig struct {
	Voice VoiceConfig `yaml:"voice" json:"voice"`
	// TODO(nullswan): Differentiate between real-time voice and alway-on voice
	// TODO(nullswan): Add video input
	// TODO(nullswan): Add image input
}

type OutputConfig struct {
	Sqlite SqliteConfig `yaml:"sqlite" json:"sqlite"`
	Speech SpeechConfig `yaml:"speech" json:"speech"`
}

type VoiceConfig struct {
	Enabled  bool   `yaml:"enabled"  json:"enabled"`
	Language string `yaml:"language" json:"language"`
	KeyCode  uint16 `yaml:"keyCode"  json:"keyCode"`
}

type SqliteConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Path    string `yaml:"path"    json:"path"`
}

type SpeechConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}
