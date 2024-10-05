package main

import (
	"context"
	"fmt"
	"os"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nullswan/golem/internal/chat"
	"github.com/nullswan/golem/internal/completion"
	"github.com/nullswan/golem/internal/config"
	provider "github.com/nullswan/golem/internal/providers/base"
	olamalocalprovider "github.com/nullswan/golem/internal/providers/ollamalocalprovider"
	"github.com/nullswan/golem/internal/providers/openaiprovider"
	"github.com/nullswan/golem/internal/ui"

	prompts "github.com/nullswan/golem/internal/prompt"
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
	Short: "An enhanced AI runtime, focusing on ease of use and extensibility.",
	Run: func(cmd *cobra.Command, args []string) {
		selectedPrompt := &prompts.DefaultPrompt
		if prompt != "" {
			var err error
			selectedPrompt, err = prompts.LoadPrompt(prompt)
			if err != nil {
				fmt.Printf("Error loading prompt: %v\n", err)
				os.Exit(1)
			}
		}

		commandCh := make(chan string)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Initialize model with channels
		model := ui.NewModel(commandCh)

		program := tea.NewProgram(
			model,
			tea.WithAltScreen(),       // Use the terminal's alternate screen
			tea.WithMouseCellMotion(), // Enable mouse events
		)

		textToTextBackend := initializeTextToTextProvider()
		conversation := chat.NewStackedConversation()
		conversation.WithPrompt(*selectedPrompt)

		go handleCommands(
			ctx,
			commandCh,
			conversation,
			textToTextBackend,
			program,
		)

		_, err := program.Run()
		if err != nil {
			os.Exit(1)
		}

		cancel()
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
	// rootCmd.AddCommand(outputCmd)
	// outputCmd.AddCommand(outputListCmd)
	// outputCmd.AddCommand(outputAddCmd)
	// #endregion

	// #region Plugin commands
	// rootCmd.AddCommand(pluginCmd)
	// pluginCmd.AddCommand(pluginListCmd)
	// pluginCmd.AddCommand(pluginEnableCmd)
	// pluginCmd.AddCommand(pluginDisableCmd)
	// #endregion

	// #region Update commands
	// rootCmd.AddCommand(updateCmd)
	// #endregion

	// #region Prompt commands
	rootCmd.AddCommand(promptCmd)
	promptCmd.AddCommand(promptListCmd)
	promptCmd.AddCommand(promptAddCmd)
	// #endregion

	// Attach flags to rootCmd only, so they are not inherited by subcommands
	rootCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Specify a prompt")
	rootCmd.Flags().
		StringVarP(&conversationID, "conversation", "c", "", "Specify a conversation ID")

	// Initialize cfg in PersistentPreRun, making it available to all commands
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if !config.ConfigExists() {
			fmt.Println("Looks like this is your first time running Golem! 🗿")
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

func initializeTextToTextProvider() provider.TextToTextProvider {
	// Check for OpenAI API key
	if os.Getenv("OPENAI_API_KEY") != "" {
		oaiConfig := openaiprovider.NewOAIProviderConfig(
			os.Getenv("OPENAI_API_KEY"),
			"",
		)
		return openaiprovider.NewTextToTextProvider(
			oaiConfig,
		)
	}

	// Default to OLama local provider
	ollamaConfig := olamalocalprovider.NewOlamaProviderConfig(
		"http://localhost:11434",
		"",
	)
	return olamalocalprovider.NewTextToTextProvider(
		ollamaConfig,
	)
}

func handleCommands(
	ctx context.Context,
	commandCh chan string,
	conversation chat.Conversation,
	textToTextBackend provider.TextToTextProvider,
	program *tea.Program,
) {
	currentCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return
		case text := <-commandCh:
			cancel()

			currentCtx, cancel = context.WithCancel(ctx)
			defer cancel()

			program.Send(ui.NewPagerMsg(text, ui.Human))
			conversation.AddMessage(chat.NewMessage(chat.RoleUser, text))

			go generateCompletion(
				currentCtx,
				conversation,
				textToTextBackend,
				program,
			)
		}
	}
}

func generateCompletion(
	ctx context.Context,
	conversation chat.Conversation,
	textToTextBackend provider.TextToTextProvider,
	program *tea.Program,
) {
	outCh := make(chan completion.Completion)

	go func() {
		defer close(outCh)
		err := textToTextBackend.GenerateCompletion(
			ctx,
			conversation.GetMessages(),
			outCh,
		)
		if err != nil {
			fmt.Printf("Error generating completion: %v\n", err)
		}
	}()

	for {
		select {
		case cmpl, ok := <-outCh:
			if reflect.TypeOf(
				cmpl,
			) == reflect.TypeOf(
				completion.CompletionTombStone{},
			) {
				program.Send(ui.NewPagerMsg("", ui.AI).WithStop())
				conversation.AddMessage(
					chat.NewMessage(chat.RoleAssistant, cmpl.Content()),
				)

				return
			}

			if !ok {
				return
			}

			program.Send(ui.NewPagerMsg(cmpl.Content(), ui.AI))
		case <-ctx.Done():
			return
		}
	}
}
