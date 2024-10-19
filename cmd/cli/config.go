package main

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v2"

	"github.com/nullswan/nomi/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := yaml.Marshal(cfg)
		if err != nil {
			log.Fatalf("Error marshalling config: %v", err)
		}

		fmt.Println(string(data))
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration parameter",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		var err error
		err = config.SetConfigValue(cfg, key, value)
		if err != nil {
			log.Fatalf("Error setting configuration: %v", err)
		}

		err = config.SaveConfig(cfg)
		if err != nil {
			log.Fatalf("Error saving configuration: %v", err)
		}

		fmt.Printf("Configuration '%s' set to '%s'\n", key, value)
	},
}

var configSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up initial configuration",
	Run: func(cmd *cobra.Command, args []string) {
		err := config.Setup()
		if err != nil {
			log.Fatalf("Error during setup: %v", err)
		}
	},
}
