package config

type Config struct {
	Input  InputConfig  `yaml:"input"  json:"input"`
	Output OutputConfig `yaml:"output" json:"output"`
	// TODO(nullswan): Add memory configuration
}

// Manage the input sources
type InputConfig struct {
	Voice EnabledConfig `yaml:"voice" json:"voice"`
	// TODO(nullswan): Differentiate between real-tme voice and alway-on voice
	// TODO(nullswan): Add video input
	// TODO(nullswan): Add image input
}

type OutputConfig struct {
	Markdown OutputDetailConfig `yaml:"markdown" json:"markdown"`
	Sqlite   OutputDetailConfig `yaml:"sqlite"   json:"sqlite"`
}

type EnabledConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type OutputDetailConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Path    string `yaml:"path"    json:"path"`
}
