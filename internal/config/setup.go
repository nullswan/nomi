package config

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/nullswan/nomi/internal/term"
)

func Setup() error {
	cfg := defaultConfig()
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

	if err := SaveConfig(&cfg); err != nil {
		return err
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

// TODO(nullswan): Create a bit of resiliency in the setup process.
// Ideally, the file could be wrapped directly to the memory.
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
		// Exec nomi prompt add <url>
		fmt.Printf("Adding prompt from %s\n", url)

		command := "nomi"
		args := []string{"prompt", "add", url}

		// Create a new process
		cmd := exec.Command(command, args...)

		// Set the output to the current stdout
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Run the command
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error adding prompt: %v\n", err)
		}
	}
}
