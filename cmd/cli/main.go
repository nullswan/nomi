package main

import (
	"fmt"
	"os"

	"github.com/nullswan/ai/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfg            *config.Config
	prompt         string
	conversationID string
)

const (
	binName = "golem"
)

var rootCmd = &cobra.Command{
	Use:   binName + " [flags] [arguments]",
	Short: binName + " is an AI runtime",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting application...")

		if prompt != "" {
			fmt.Printf("Using prompt: %s\n", prompt)
			// TODO: Load and apply the specified prompt
		} else {
			fmt.Println("Using default prompt")
			// TODO: Use the default prompt
		}

		if conversationID != "" {
			fmt.Printf("Resuming conversation ID: %s\n", conversationID)
			// TODO: Load and resume the specified conversation
		} else {
			fmt.Println("Creating new conversation")
			// TODO: Create and load the default conversation
		}

		// TODO: Start the main application logic
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func main() {
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

	// #region Prompt commands
	rootCmd.AddCommand(promptCmd)
	promptCmd.AddCommand(promptListCmd)
	// #endregion

	// Attach flags to rootCmd only, so they are not inherited by subcommands
	rootCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Specify a prompt")
	rootCmd.Flags().
		StringVarP(&conversationID, "conversation", "c", "", "Specify a conversation ID")

	// Initialize cfg in PersistentPreRun, making it available to all commands
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if !config.ConfigExists() {
			if err := config.Setup(); err != nil {
				fmt.Printf("Error during configuration setup: %v\n", err)
				os.Exit(1)
			}
		}

		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
	}

	// Execute the root command
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
