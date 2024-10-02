package config

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
)

func Setup() error {
	cfg := defaultConfig()
	fmt.Println("Starting configuration setup...")

	cfg.Input.Text.Enabled = promptForBool(
		"Enable text input",
		cfg.Input.Text.Enabled,
	)
	cfg.Input.Voice.Enabled = promptForBool(
		"Enable voice input",
		cfg.Input.Voice.Enabled,
	)

	cfg.Output.Markdown.Enabled = promptForBool(
		"Enable Markdown output",
		cfg.Output.Markdown.Enabled,
	)
	if cfg.Output.Markdown.Enabled {
		cfg.Output.Markdown.Path = promptForString(
			"Path for Markdown output",
			cfg.Output.Markdown.Path,
		)
	}

	cfg.Output.Sqlite.Enabled = promptForBool(
		"Enable SQLite output",
		cfg.Output.Sqlite.Enabled,
	)
	if cfg.Output.Sqlite.Enabled {
		cfg.Output.Sqlite.Path = promptForString(
			"Path for SQLite output",
			cfg.Output.Sqlite.Path,
		)
	}

	memoryModes := []MemoryMode{
		MemoryModeConversation,
		MemoryModeNone,
	}
	memoryModeLabels := []string{
		"Knowledge Graph",
		"Entity",
		"Conversation",
		"None",
	}
	modeIndex := promptForSelect(
		"Choose memory mode",
		memoryModeLabels,
		getDefaultMemoryModeIndex(cfg.Memory.Mode),
	)
	cfg.Memory.Mode = memoryModes[modeIndex]

	if err := SaveConfig(&cfg); err != nil {
		return err
	}

	fmt.Println("Configuration setup completed.")
	return nil
}

func promptForBool(label string, defaultVal bool) bool {
	items := []string{"Yes", "No"}
	defaultIndex := 0
	if !defaultVal {
		defaultIndex = 1
	}
	prompt := promptui.Select{
		Label:        label,
		Items:        items,
		CursorPos:    defaultIndex,
		HideHelp:     true,
		HideSelected: true,
	}
	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		os.Exit(1)
	}
	return result == "Yes"
}

func promptForString(label string, defaultVal string) string {
	prompt := promptui.Prompt{
		Label:   label,
		Default: defaultVal,
	}
	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		os.Exit(1)
	}
	return result
}

func promptForSelect(label string, items []string, defaultIndex int) int {
	prompt := promptui.Select{
		Label:        label,
		Items:        items,
		CursorPos:    defaultIndex,
		HideHelp:     true,
		HideSelected: true,
	}
	index, _, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		os.Exit(1)
	}
	return index
}

func getDefaultMemoryModeIndex(mode MemoryMode) int {
	switch mode {
	case MemoryModeConversation:
		return 2
	case MemoryModeNone:
		return 3
	default:
		return 3 // Default to 'None' if unrecognized
	}
}
