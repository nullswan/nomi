package setup

import (
	"fmt"

	"github.com/nullswan/nomi/internal/config"
	prompts "github.com/nullswan/nomi/internal/prompt"
	"github.com/nullswan/nomi/internal/term"
)

// Setup handles the configuration setup
// It prompts the user for configuration values and saves them to the configuration file
func Setup() error {
	cfg := config.DefaultConfig()
	fmt.Println("Starting configuration setup...")

	cfg.DevMode = term.PromptForBool(
		"Enable development mode",
		cfg.DevMode,
	)

	cfg.Input.Voice.Enabled = term.PromptForBool(
		"Enable voice input",
		cfg.Input.Voice.Enabled,
	)

	cfg.Output.Sqlite.Enabled = true
	if cfg.Output.Sqlite.Enabled {
		cfg.Output.Sqlite.Path = term.PromptForString(
			"Path for the SQLite database",
			cfg.Output.Sqlite.Path,
		)
	}

	if err := config.SaveConfig(&cfg); err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

	var doInstallDefaultPrompts bool
	doInstallDefaultPrompts = term.PromptForBool(
		"Install default prompts ? [Recommended]",
		doInstallDefaultPrompts,
	)

	if doInstallDefaultPrompts {
		installDefaultPrompts()
	}

	fmt.Println("Configuration setup completed.")
	return nil
}

func installDefaultPrompts() {
	fmt.Println("Installing default prompts...")
	urls := []string{
		"https://raw.githubusercontent.com/nullswan/nomi/refs/heads/main/prompts/native-ask.yml",
		"https://raw.githubusercontent.com/nullswan/nomi/refs/heads/main/prompts/native-code.yml",
		"https://raw.githubusercontent.com/nullswan/nomi/refs/heads/main/prompts/native-commit-message.yml",
		"https://raw.githubusercontent.com/nullswan/nomi/refs/heads/main/prompts/native-rephrase.yml",
		"https://raw.githubusercontent.com/nullswan/nomi/refs/heads/main/prompts/native-summarize.yml",
	}

	for _, url := range urls {
		fmt.Printf("Adding prompt from %s\n", url)

		_, err := prompts.AddPromptFromURL(url)
		if err != nil {
			fmt.Printf("Error adding prompt: %v\n", err)
			continue
		}

		fmt.Println("Prompt added successfully.")
	}
}
