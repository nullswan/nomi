package main

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/nullswan/golem/internal/chat"
	"github.com/nullswan/golem/internal/completion"
	"github.com/nullswan/golem/internal/config"
	"github.com/nullswan/golem/internal/providers"
	provider "github.com/nullswan/golem/internal/providers/base"
	olamalocalprovider "github.com/nullswan/golem/internal/providers/ollamalocalprovider"
	"github.com/nullswan/golem/internal/providers/openaiprovider"
	"github.com/nullswan/golem/internal/term"

	prompts "github.com/nullswan/golem/internal/prompt"
	"github.com/spf13/cobra"
)

var (
	cfg                 *config.Config
	startPrompt         string
	interactiveMode     bool
	startConversationId string
)

const (
	binName = "golem"
)

var rootCmd = &cobra.Command{
	Use:   binName + " [flags] [arguments]",
	Short: "An enhanced AI runtime, focusing on ease of use and extensibility.",
	Run: func(cmd *cobra.Command, args []string) {
		selectedPrompt := &prompts.DefaultPrompt
		if startPrompt != "" {
			var err error
			selectedPrompt, err = prompts.LoadPrompt(startPrompt)
			if err != nil {
				fmt.Printf("Error loading prompt: %v\n", err)
				os.Exit(1)
			}
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		provider := providers.FindFirstProvider()
		textToTextBackend := initializeTextToTextProvider()

		// TODO(nullswan): Handle restarting conversation
		conversation := chat.NewStackedConversation()
		conversation.WithPrompt(*selectedPrompt)

		// Welcome message
		fmt.Printf("----\n")
		fmt.Printf("Welcome to Golem! 🗿\n")
		fmt.Println()
		fmt.Println("Configuration")
		fmt.Printf("  Start prompt: %s\n", startPrompt)
		fmt.Printf("  Conversation: %s\n", conversation.GetId())
		fmt.Printf("  Provider: %s\n", provider)
		fmt.Printf("  Build version: %s %s\n", buildVersion, buildDate)
		fmt.Printf("-----\n")

		pipedInput, err := term.GetPipedInput()
		if err != nil {
			fmt.Printf("Error reading piped input: %v\n", err)
		}

		var cancelRequest context.CancelFunc

		processInput := func(text string) {
			if cancelRequest != nil {
				cancelRequest()
			}

			requestContext, newCancelRequest := context.WithCancel(ctx)
			cancelRequest = newCancelRequest

			if text == "" {
				fmt.Println()
				return
			}

			fmt.Printf("You:\n%s", text)
			conversation.AddMessage(chat.NewMessage(chat.RoleUser, text))

			generateCompletion(
				requestContext,
				conversation,
				textToTextBackend,
			)
		}

		if pipedInput != "" {
			processInput(pipedInput)
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Fprint(os.Stderr, "\n")
				text := term.NewInputArea()
				processInput(text)
			}
		}
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func main() {
	// #region Config commands
	// rootCmd.AddCommand(configCmd)
	// configCmd.AddCommand(configShowCmd)
	// configCmd.AddCommand(configSetCmd)
	// configCmd.AddCommand(configSetupCmd)
	// #endregion

	// #region Conversation commands
	rootCmd.AddCommand(conversationCmd)
	conversationCmd.AddCommand(conversationListCmd)
	// #endregion

	// #region Version commands
	rootCmd.AddCommand(versionCmd)
	// #endregion

	// TODO(nullswan): Add update command

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
		StringVarP(&startConversationId, "conversation", "c", "", "Open a conversation by ID")
	rootCmd.Flags().
		BoolVarP(&interactiveMode, "interactive", "i", false, "Start in interactive mode")

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

// TODO(nullswan): Check provider validity
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

	// TODO(nullswan): Check if OLama local provider is running
	// Default to OLama local provider
	ollamaConfig := olamalocalprovider.NewOlamaProviderConfig(
		"http://localhost:11434",
		"",
	)
	return olamalocalprovider.NewTextToTextProvider(
		ollamaConfig,
	)
}

func generateCompletion(
	ctx context.Context,
	conversation chat.Conversation,
	textToTextBackend provider.TextToTextProvider,
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
			// TODO(nullswan): Use logger
			// fmt.Printf("Error generating completion: %v\n", err)
		}
	}()

	fmt.Printf("AI: \n")
	for {
		select {
		case cmpl, ok := <-outCh:
			if reflect.TypeOf(
				cmpl,
			) == reflect.TypeOf(
				completion.CompletionTombStone{},
			) {
				fmt.Println()
				// program.Send(ui.NewPagerMsg("", ui.AI).WithStop())
				conversation.AddMessage(
					chat.NewMessage(chat.RoleAssistant, cmpl.Content()),
				)

				return
			}

			if !ok {
				return
			}

			fmt.Printf("%s", cmpl.Content())
			// program.Send(ui.NewPagerMsg(cmpl.Content(), ui.AI))
		case <-ctx.Done():
			return
		}
	}
}
