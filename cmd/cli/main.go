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
		} else {
			startPrompt = "default"
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		provider := providers.CheckProvider()
		textToTextBackend, err := providers.LoadTextToTextProvider(provider, "")
		if err != nil {
			fmt.Printf("Error loading text-to-text provider: %v\n", err)
			os.Exit(1)
		}
		defer textToTextBackend.Close()

		repo, err := chat.NewSQLiteRepository(cfg.Output.Sqlite.Path)
		if err != nil {
			fmt.Printf("Error creating repository: %v\n", err)
			os.Exit(1)
		}
		defer repo.Close()

		var conversation chat.Conversation
		if startConversationId != "" {
			conversation, err = repo.LoadConversation(startConversationId)
			if err != nil {
				fmt.Printf("Error loading conversation: %v\n", err)
				os.Exit(1)
			}
		} else {
			conversation = chat.NewStackedConversation(repo)
			conversation.WithPrompt(*selectedPrompt)
		}

		// Welcome message
		if !interactiveMode {
			fmt.Printf("----\n")
			fmt.Printf("Welcome to Golem! (v%s) 🗿\n", buildVersion)
			fmt.Println()
			fmt.Println("Configuration")
			fmt.Printf("  Start prompt: %s\n", startPrompt)
			fmt.Printf("  Conversation: %s\n", conversation.GetId())
			fmt.Printf("  Provider: %s\n", provider)
			fmt.Printf("  Build Date: %s\n", buildDate)
			fmt.Printf("-----\n")
			fmt.Printf("Press Enter twice to send a message.\n")
			fmt.Printf("Press Ctrl+C to exit.\n")
			fmt.Printf("Press Ctrl+K to cancel the current request.\n")
			fmt.Printf("-----\n\n")
		}

		fmt.Printf("EXIT")
		os.Exit(1)

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

			completion, err := generateCompletion(
				requestContext,
				conversation,
				textToTextBackend,
				interactiveMode,
			)
			if err != nil {
				fmt.Printf("Error generating completion: %v\n", err)
				return
			}

			conversation.AddMessage(
				chat.NewMessage(chat.RoleAssistant, completion),
			)
		}

		if pipedInput != "" {
			processInput(pipedInput)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
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

func generateCompletion(
	ctx context.Context,
	conversation chat.Conversation,
	textToTextBackend provider.TextToTextProvider,
	useRender bool,
) (string, error) {
	outCh := make(chan completion.Completion)

	go func() {
		defer close(outCh)
		if err := textToTextBackend.GenerateCompletion(ctx, conversation.GetMessages(), outCh); err != nil {
			fmt.Printf("Error generating completion: %v\n", err)
		}
	}()

	fmt.Println("AI:")
	var fullContent string

	for {
		select {
		case cmpl, ok := <-outCh:
			if isTombStone(cmpl) {
				fmt.Println()
				return fullContent, nil
			}

			if !ok {
				fmt.Println()
				return fullContent, fmt.Errorf("error reading completion")
			}

			if cmpl.Content() == "" {
				continue
			}

			fullContent += cmpl.Content()
			fmt.Printf("%s", cmpl.Content())
		case <-ctx.Done():
			return fullContent, fmt.Errorf("context canceled")
		}
	}
}

func isTombStone(cmpl completion.Completion) bool {
	return reflect.TypeOf(
		cmpl,
	) == reflect.TypeOf(
		completion.CompletionTombStone{},
	)
}
