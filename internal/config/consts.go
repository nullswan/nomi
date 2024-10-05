package config

const (
	defaultConversationDir = "conversations"

	// configDir is the directory where the configuration file is stored.
	configDir = ".golem"

	// configFileName is the name of the configuration file.
	configFileName = "config.yml"
)

var configFilePath string

func init() {
	configFilePath = getConfigFilePath()
}
