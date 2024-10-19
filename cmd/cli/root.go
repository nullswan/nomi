package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/nullswan/nomi/internal/config"
	"github.com/spf13/cobra"
)

const ErrLocalWhisperNotSupported = "OPENAI_API_KEY is not set, voice input will be disabled -- local whisper will be supported soon!"

func main() {
	// #region Config commands
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configSetupCmd)
	// #endregion

	// #region Interpreter commands
	rootCmd.AddCommand(interpreterCmd)
	// #endregion

	// #region Conversation commands
	rootCmd.AddCommand(conversationCmd)
	conversationCmd.AddCommand(conversationListCmd)
	conversationCmd.AddCommand(conversationShowCmd)
	conversationCmd.AddCommand(conversationDeleteCmd)
	// #endregion

	// #region Version commands
	rootCmd.AddCommand(versionCmd)
	// #endregion

	// #region Update commands
	rootCmd.AddCommand(updateCmd)
	// #endregion

	// #region Prompt commands
	rootCmd.AddCommand(promptCmd)
	promptCmd.AddCommand(promptListCmd)
	promptCmd.AddCommand(promptAddCmd)
	promptCmd.AddCommand(promptEditCmd)
	// #endregion

	// Attach flags to rootCmd only, so they are not inherited by subcommands
	rootCmd.Flags().
		StringVarP(&startPrompt, "prompt", "p", "", "Specify a prompt")
	rootCmd.Flags().
		StringVarP(&targetModel, "model", "m", "", "Specify a model")
	rootCmd.Flags().
		StringVarP(&startConversationID, "conversation", "c", "", "Open a conversation by ID")
	rootCmd.Flags().
		BoolVarP(&interactiveMode, "interactive", "i", false, "Start in interactive mode")

	// Initialize cfg in PersistentPreRun, making it available to all commands
	rootCmd.PersistentPreRun = func(_ *cobra.Command, _ []string) {
		if !config.Exists() {
			fmt.Println("Looks like this is your first time running nomi!")
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

		oaiKey := os.Getenv("OPENAI_API_KEY")
		if oaiKey == "" && cfg.Input.Voice.Enabled {
			fmt.Println(
				ErrLocalWhisperNotSupported,
			)
			cfg.Input.Voice.Enabled = false
		}

		if cfg.DevMode {
			go func() {
				if err := http.ListenAndServe("localhost:6060", nil); err != nil {
					fmt.Printf("Error starting pprof server: %v\n", err)
					os.Exit(1)
				}
			}()
		}
	}

	// Execute the root command
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
