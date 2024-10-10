package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"

	"github.com/charmbracelet/glamour"
	"github.com/manifoldco/promptui/screenbuf"
	"github.com/nullswan/golem/internal/chat"
	"github.com/nullswan/golem/internal/completion"
	"github.com/nullswan/golem/internal/config"
	"github.com/nullswan/golem/internal/providers"
	baseprovider "github.com/nullswan/golem/internal/providers/base"
	"github.com/nullswan/golem/internal/term"

	prompts "github.com/nullswan/golem/internal/prompt"
	"github.com/spf13/cobra"
)

var (
	cfg                 *config.Config
	startPrompt         string
	interactiveMode     bool
	startConversationID string
	targetModel         string
)

const (
	binName = "golem"
)

var rootCmd = &cobra.Command{
	Use:   binName + " [flags] [arguments]",
	Short: "An enhanced AI runtime, focusing on ease of use and extensibility.",
	Run: func(_ *cobra.Command, args []string) {
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
		textToTextBackend, err := providers.LoadTextToTextProvider(
			provider,
			targetModel,
		)
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
		if startConversationID != "" {
			conversation, err = repo.LoadConversation(startConversationID)
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
			fmt.Printf("Welcome to Golem! (%s) 🗿\n", buildVersion)
			fmt.Println()
			fmt.Println("Configuration")
			fmt.Printf("  Start prompt: %s\n", startPrompt)
			fmt.Printf("  Conversation: %s\n", conversation.GetID())
			fmt.Printf("  Provider: %s\n", provider)
			fmt.Printf("  Build Date: %s\n", buildDate)
			fmt.Printf("-----\n")
			fmt.Printf("Press Enter twice to send a message.\n")
			fmt.Printf("Press Ctrl+C to exit.\n")
			fmt.Printf("Press Ctrl+K to cancel the current request.\n")
			fmt.Printf("-----\n\n")
		}

		pipedInput, err := term.GetPipedInput()
		if err != nil {
			fmt.Printf("Error reading piped input: %v\n", err)
		}

		renderer, err := term.InitRenderer()
		if err != nil {
			fmt.Printf("Error initializing renderer: %v\n", err)
			os.Exit(1)
		}

		var cancelRequest context.CancelFunc
		processInput := func(text string) {
			if cancelRequest != nil {
				cancelRequest()
			}

			requestContext, newCancelRequest := context.WithCancel(ctx)
			cancelRequest = newCancelRequest

			if text == "" {
				return
			}

			if isLocalResource(text) {
				processLocalResource(conversation, text)
				return
			}

			conversation.AddMessage(chat.NewMessage(chat.RoleUser, text))

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigCh
				cancelRequest()
			}()
			defer signal.Stop(sigCh)
			defer close(sigCh)

			completion, err := generateCompletion(
				requestContext,
				conversation,
				textToTextBackend,
				renderer,
			)
			if err != nil {
				if strings.Contains(err.Error(), "context canceled") {
					fmt.Println("\nRequest canceled by the user.")
					return
				}
				fmt.Printf("Error generating completion: %v\n", err)
				return
			}

			conversation.AddMessage(
				chat.NewMessage(chat.RoleAssistant, completion),
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
				text, err := term.NewInputArea()
				if err != nil {
					if errors.Is(err, term.ErrInputInterrupted) ||
						errors.Is(err, term.ErrInputKilled) {
						cancel()
						return
					}
					fmt.Printf("Error reading input: %v\n", err)
					return
				}
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
	conversationCmd.AddCommand(conversationShowCmd)
	conversationCmd.AddCommand(conversationDeleteCmd)
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
		StringVarP(&targetModel, "model", "m", "", "Specify a model")
	rootCmd.Flags().
		StringVarP(&startConversationID, "conversation", "c", "", "Open a conversation by ID")
	rootCmd.Flags().
		BoolVarP(&interactiveMode, "interactive", "i", false, "Start in interactive mode")

	// Initialize cfg in PersistentPreRun, making it available to all commands
	rootCmd.PersistentPreRun = func(_ *cobra.Command, args []string) {
		if !config.Exists() {
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
	textToTextBackend baseprovider.TextToTextProvider,
	renderer *glamour.TermRenderer,
) (string, error) {
	outCh := make(chan completion.Completion)

	go func() {
		defer close(outCh)
		if err := textToTextBackend.GenerateCompletion(ctx, conversation.GetMessages(), outCh); err != nil {
			if strings.Contains(err.Error(), "context canceled") {
				return
			}
			fmt.Printf("Error generating completion: %v\n", err)
		}
	}()

	sb := screenbuf.New(os.Stdout)
	var fullContent string
	currentLine := ""

	for {
		select {
		case cmpl, ok := <-outCh:
			if isTombStone(cmpl) {
				sb.Clear()
				fullContent += currentLine
				currentLine = ""

				mdContent, err := renderer.Render(fullContent)
				if err != nil {
					fmt.Println("Error rendering markdown:", err)
					return fullContent, fmt.Errorf(
						"rendering markdown: %w",
						err,
					)
				}
				fmt.Println(mdContent)
				return fullContent, nil
			}

			if !ok {
				fmt.Println()
				return fullContent, errors.New("error reading completion")
			}

			if cmpl.Content() == "" {
				continue
			}

			fullContent += cmpl.Content()
			currentLine += cmpl.Content()
			if strings.Contains(currentLine, "\n") {
				sb.WriteString(currentLine)
				currentLine = currentLine[strings.LastIndex(currentLine, "\n")+1:]
			}
		case <-ctx.Done():
			return fullContent, errors.New("context canceled")
		}
	}
}

func isTombStone(cmpl completion.Completion) bool {
	return reflect.TypeOf(
		cmpl,
	) == reflect.TypeOf(
		completion.Tombstone{},
	)
}

func isLocalResource(text string) bool {
	var path string
	if strings.HasPrefix(text, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		path = filepath.Join(home, text[1:])
	} else {
		path = text
	}

	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "/") {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

func processLocalResource(conversation chat.Conversation, text string) {
	if isDirectory(text) {
		addAllFiles(conversation, text)
	} else {
		addSingleFile(conversation, text)
	}
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func addAllFiles(conversation chat.Conversation, directory string) {
	files, err := os.ReadDir(directory)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}
	for _, file := range files {
		path := filepath.Join(directory, file.Name())
		if file.IsDir() {
			addAllFiles(conversation, path)
		} else {
			addFileToConversation(conversation, path, file.Name())
		}
	}
}

func addSingleFile(conversation chat.Conversation, filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	fileName := filepath.Base(filePath)
	conversation.AddMessage(
		chat.NewFileMessage(
			chat.RoleUser,
			formatFileMessage(fileName, string(content)),
		),
	)
}

func addFileToConversation(
	conversation chat.Conversation,
	filePath, fileName string,
) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	conversation.AddMessage(
		chat.NewFileMessage(
			chat.RoleUser,
			formatFileMessage(fileName, string(content)),
		),
	)

	fmt.Printf("Added file: %s\n", filePath)
}

func formatFileMessage(fileName, content string) string {
	return fileName + "-----\n" + content + "-----\n"
}
