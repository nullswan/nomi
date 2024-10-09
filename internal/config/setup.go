package config

// TODO(nullswan): Use bubbletea for the setup process instead of promptui.
import (
	"fmt"

	"github.com/nullswan/golem/internal/term"
)

func Setup() error {
	cfg := defaultConfig()
	fmt.Println("Starting configuration setup...")

	cfg.Input.Voice.Enabled = term.PromptForBool(
		"Enable voice input",
		cfg.Input.Voice.Enabled,
	)

	cfg.Output.Markdown.Enabled = term.PromptForBool(
		"Enable Markdown output",
		cfg.Output.Markdown.Enabled,
	)
	if cfg.Output.Markdown.Enabled {
		cfg.Output.Markdown.Path = term.PromptForString(
			"Path for Markdown output",
			cfg.Output.Markdown.Path,
		)
	}

	cfg.Output.Sqlite.Enabled = term.PromptForBool(
		"Enable SQLite output",
		cfg.Output.Sqlite.Enabled,
	)
	if cfg.Output.Sqlite.Enabled {
		cfg.Output.Sqlite.Path = term.PromptForString(
			"Path for SQLite output",
			cfg.Output.Sqlite.Path,
		)
	}

	if err := SaveConfig(&cfg); err != nil {
		return err
	}

	fmt.Println("Configuration setup completed.")
	return nil
}
