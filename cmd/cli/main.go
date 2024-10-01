package main

import (
	"fmt"
	"os"

	"github.com/nullswan/ai/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ai [flags] [arguments]",
	Short: "YourApp is an AI assistant CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting application...")

		// If config is not set, create config.

		// if preset != "" {
		// 	fmt.Printf("Using preset: %s\n", preset)
		// 	// TODO: Load and apply the specified preset
		// } else {
		// 	fmt.Println("No preset specified.")
		// 	// TODO: Use the default preset
		// }

		// if conversationID != "" {
		// 	fmt.Printf("Resuming conversation ID: %s\n", conversationID)
		// 	// TODO: Load and resume the specified conversation
		// } else {
		// 	fmt.Println("No conversation ID specified.")
		// 	// TODO: Load the default conversation
		// }

		// // TODO: Start the main application logic
	},
}

func main() {
	if !config.ConfigExists() {
		if err := config.Setup(); err != nil {
			fmt.Printf("Error during configuration setup: %v\n", err)
			os.Exit(1)
		}
	}

	// #region Config commands
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configSetupCmd)
	// #endregion

	// #region Conversation commands
	rootCmd.AddCommand(conversationCmd)
	conversationCmd.AddCommand(conversationListCmd)
	// #endregion

	// #region Output commands
	rootCmd.AddCommand(outputCmd)
	outputCmd.AddCommand(outputListCmd)
	outputCmd.AddCommand(outputAddCmd)
	// #endregion

	// #region Plugin commands
	rootCmd.AddCommand(pluginCmd)
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginEnableCmd)
	pluginCmd.AddCommand(pluginDisableCmd)
	// #endregion

	// #region Preset commands
	rootCmd.AddCommand(presetCmd)
	presetCmd.AddCommand(presetListCmd)
	// #endregion

	var preset, conversationID string

	rootCmd.PersistentFlags().
		StringVarP(&preset, "preset", "p", "", "Specify a preset")
	rootCmd.PersistentFlags().
		StringVarP(&conversationID, "conversation", "c", "", "Specify a conversation ID")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
