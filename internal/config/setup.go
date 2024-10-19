package config

// TODO(nullswan): Use bubbletea for the setup process instead of promptui.
import (
	"fmt"

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

	fmt.Println("Configuration setup completed.")
	return nil
}
