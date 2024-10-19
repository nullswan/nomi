package config

type Config struct {
	Input   InputConfig  `yaml:"input"    json:"input"`
	Output  OutputConfig `yaml:"output"   json:"output"`
	DevMode bool         `yaml:"dev_mode" json:"dev_mode"`
	// TODO(nullswan): Add memory configuration
}

// Manage the input sources
type InputConfig struct {
	Voice EnabledConfig `yaml:"voice" json:"voice"`
	// TODO(nullswan): Differentiate between real-time voice and alway-on voice
	// TODO(nullswan): Add video input
	// TODO(nullswan): Add image input
}

type OutputConfig struct {
	Sqlite OutputDetailConfig `yaml:"sqlite" json:"sqlite"`
}

type EnabledConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type OutputDetailConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Path    string `yaml:"path"    json:"path"`
}
