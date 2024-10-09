package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
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
	targetModel         string
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

		pipedInput, err := term.GetPipedInput()
		if err != nil {
			fmt.Printf("Error reading piped input: %v\n", err)
		}

		var cancelRequest context.CancelFunc

		var renderer *glamour.TermRenderer
		if !interactiveMode {
			var styleOpt glamour.TermRendererOption
			darkStyle := styles.DarkStyleConfig
			darkStyle.Document.Margin = uintPtr(0)
			darkStyle.CodeBlock.Margin = uintPtr(0)
			styleOpt = glamour.WithStyles(darkStyle)

			renderer, err = glamour.NewTermRenderer(
				styleOpt,
				glamour.WithWordWrap(0),
				glamour.WithEmoji(),
			)
			if err != nil {
				fmt.Printf("Error creating renderer: %v\n", err)
				return
			}
		}

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

			if isLocalResource(text) {
				processLocalResource(conversation, text)
				return
			}

			if interactiveMode {
				fmt.Printf("You:\n%s\n\n", text)
			} else {
				style := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FF00FF"))
				fmt.Printf("%s\n", style.Render("You:"))
				renderedContent, err := renderer.Render(text)
				if err != nil {
					fmt.Printf("Error rendering input: %v\n", err)
					return
				}

				fmt.Printf("%s\n", renderedContent)
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
					fmt.Println("Request canceled.")
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
		StringVarP(&targetModel, "model", "m", "", "Specify a model")
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

	fmt.Println("AI:")
	var fullContent string

	for {
		select {
		case cmpl, ok := <-outCh:
			if isTombStone(cmpl) {
				if renderer != nil {
					renderedContent, err := renderer.Render(fullContent)
					if err != nil {
						fmt.Printf("Error rendering completion: %v\n", err)
						return fullContent, err
					}

					lines := strings.Split(fullContent, "\n")
					fmt.Print("\033[2K")
					for i := 0; i < len(lines)+1; i++ { // Need to clear AI prefix too
						fmt.Print("\033[1F\033[2K")
					}

					style := lipgloss.NewStyle().
						Foreground(lipgloss.Color("#00C0FF"))
					fmt.Printf("%s\n%s", style.Render("AI:"), renderedContent)
				}

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
	return fileName + "-----\n" + string(content) + "-----\n"
}

func uintPtr(u uint) *uint { return &u }
