package config

type Config struct {
	Input  InputConfig  `yaml:"input"  json:"input"`
	Output OutputConfig `yaml:"output" json:"output"`
	Memory MemoryConfig `yaml:"memory" json:"memory"`
}

type InputConfig struct {
	Text  EnabledConfig `yaml:"text"  json:"text"`
	Image EnabledConfig `yaml:"image" json:"image"`
	Voice EnabledConfig `yaml:"voice" json:"voice"`
}

type OutputConfig struct {
	Markdown OutputDetailConfig `yaml:"markdown" json:"markdown"`
	Sqlite   OutputDetailConfig `yaml:"sqlite"   json:"sqlite"`
}

type MemoryConfig struct {
	Mode MemoryMode `yaml:"mode" json:"mode"`
}

type EnabledConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type OutputDetailConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Path    string `yaml:"path"    json:"path"`
}

type MemoryMode string

const (
	MemoryModeKnowledgeGraph MemoryMode = "knowledge_graph"
	MemoryModeEntity         MemoryMode = "entity"
	MemoryModeConversation   MemoryMode = "conversation"
	MemoryModeNone           MemoryMode = "none"
)

func (m MemoryMode) String() string {
	return string(m)
}
