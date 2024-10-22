package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"

	"github.com/nullswan/nomi/internal/config"
	"github.com/nullswan/nomi/internal/setup"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Run: func(cmd *cobra.Command, _ []string) {
		err := cmd.Help()
		if err != nil {
			log.Fatalf("Error displaying help: %v", err)
		}
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(_ *cobra.Command, _ []string) {
		data, err := yaml.Marshal(cfg)
		if err != nil {
			log.Fatalf("Error marshalling config: %v", err)
		}

		fmt.Println(string(data))
	},
}

// TODO(nullswan): Replace with editor, just like with the prompt edit
var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration",
	Run: func(_ *cobra.Command, _ []string) {
		// Open editor for further editing
		tempFile, err := os.CreateTemp("/tmp", "*.yaml")
		if err != nil {
			log.Fatalf("Error creating temporary file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		configYaml, err := yaml.Marshal(cfg)
		if err != nil {
			log.Fatalf("Error marshalling config to YAML: %v", err)
		}

		if _, err := tempFile.Write(configYaml); err != nil {
			log.Fatalf("Error writing to temp file: %v", err)
		}
		tempFile.Close()

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		process := exec.Command(editor, tempFile.Name())
		process.Stdin = os.Stdin
		process.Stdout = os.Stdout
		process.Stderr = os.Stderr

		if err := process.Run(); err != nil {
			log.Fatalf("Error opening editor: %v", err)
		}

		updatedData, err := os.ReadFile(tempFile.Name())
		if err != nil {
			log.Fatalf("Error reading updated file: %v", err)
		}

		if err := yaml.Unmarshal(updatedData, cfg); err != nil {
			log.Fatalf("Error unmarshalling updated YAML: %v", err)
		}

		if err := config.SaveConfig(cfg); err != nil {
			log.Fatalf("Error saving updated configuration: %v", err)
		}

		fmt.Println("Configuration updated successfully")
	},
}

var configSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up initial configuration",
	Run: func(_ *cobra.Command, _ []string) {
		err := setup.Setup()
		if err != nil {
			log.Fatalf("Error during setup: %v", err)
		}
	},
}
