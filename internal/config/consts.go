package config

const (
	conversationDir = "conversations"
	promptDir       = "prompts"

	// configDir is the directory where the configuration file is stored.
	configDir = ".nomi"

	// configFileName is the name of the configuration file.
	configFileName = "config.yml"
)

var configFilePath string

func init() {
	configFilePath = getConfigFilePath()
}

func GetPromptDirectory() string {
	return GetModuleDirectory(promptDir)
}

func GetConversationDirectory() string {
	return GetModuleDirectory(conversationDir)
}
