package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nullswan/golem/internal/chat"
	"github.com/nullswan/golem/internal/code"
	prompts "github.com/nullswan/golem/internal/prompt"
	"github.com/nullswan/golem/internal/providers"
	baseprovider "github.com/nullswan/golem/internal/providers/base"
	"github.com/nullswan/golem/internal/term"
	"github.com/spf13/cobra"
)

const (
	interpreterMaxRetries = 3
)

// Add the interpreter command
var interpreterCmd = &cobra.Command{
	Use:   "interpreter",
	Short: "Start the interpreter",
	Run: func(_ *cobra.Command, _ []string) {
		// Ensure the provider is OpenAI
		provider := providers.CheckProvider()
		if provider != providers.OpenAIProvider {
			fmt.Println("Error: only openai is supported currently.")
			os.Exit(1)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		provider = providers.CheckProvider()

		var err error
		var textToTextBackend baseprovider.TextToTextProvider
		if prompts.DefaultInterpreterPrompt.Preferences.Reasoning {
			textToTextBackend, err = providers.LoadTextToTextReasoningProvider(
				provider,
				targetModel,
			)
			if err != nil {
				fmt.Printf(
					"Error loading text-to-text reasoning provider: %v\n",
					err,
				)
			}
		}
		if textToTextBackend == nil {
			textToTextBackend, err = providers.LoadTextToTextProvider(
				provider,
				targetModel,
			)
			if err != nil {
				fmt.Printf("Error loading text-to-text provider: %v\n", err)
				os.Exit(1)
			}
		}

		defer textToTextBackend.Close()

		repo, err := chat.NewSQLiteRepository(cfg.Output.Sqlite.Path)
		if err != nil {
			fmt.Printf("Error creating repository: %v\n", err)
			os.Exit(1)
		}
		defer repo.Close()

		conversation := chat.NewStackedConversation(repo)
		conversation.WithPrompt(prompts.DefaultInterpreterPrompt)

		// Display welcome message
		fmt.Printf("----\n")
		fmt.Printf("✨ Welcome to Golem Interpreter! 🗿✨\n")
		fmt.Println()
		fmt.Println("Configuration")
		fmt.Printf(
			"  Start prompt: default-interpreter\n",
		)
		fmt.Printf("  Conversation: %s\n", conversation.GetID())
		fmt.Printf("  Provider: %s\n", provider)
		fmt.Printf("  Model: %s\n", textToTextBackend.GetModel())
		fmt.Printf("  Build Date: %s\n", buildDate)
		fmt.Printf("-----\n")
		fmt.Printf("Press Enter twice to send a message.\n")
		fmt.Printf("Press Ctrl+C to exit.\n")
		fmt.Printf("Press Ctrl+K to cancel the current request.\n")
		fmt.Printf("-----\n\n")

		pipedInput, err := term.GetPipedInput()
		if err != nil {
			fmt.Printf("Error reading piped input: %v\n", err)
		}

		renderer, err := term.InitRenderer()
		if err != nil {
			fmt.Printf("Error initializing renderer: %v\n", err)
			os.Exit(1)
		}

		var lastResult []code.ExecutionResult
		var cancelRequest context.CancelFunc
		retries := 0
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

			// Exec the command
			result := code.InterpretCodeBlocks(completion)
			for _, r := range result {
				if r.ExitCode != 0 {
					fmt.Printf("Error executing command: %s\n", r.Stderr)
				}

				fmt.Printf("Received: %s\n%s\n", r.Stdout, r.Stderr)
			}

			if len(result) == 0 {
				fmt.Println("No results received.")
				return
			}

			// Extend the conversation with the result
			formattedResult := code.FormatExecutionResultForLLM(result)
			conversation.AddMessage(
				chat.NewMessage(chat.RoleAssistant, formattedResult),
			)

			lastResult = result
		}

		if pipedInput != "" {
			processInput(pipedInput)
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Loop until we get a valid input, try again if we get an error, limit retries
				if len(lastResult) > 0 {
					containsError := false
					for _, result := range lastResult {
						if result.ExitCode != 0 {
							containsError = true
							break
						}
					}

					if containsError {
						fmt.Println(
							"Error executing command, see previous output for details.",
						)

						retries++
						if retries > interpreterMaxRetries {
							fmt.Println("Max retries reached.")
							retries = 0
							lastResult = nil
						} else {
							fmt.Printf(
								"Retrying...\n",
							)
							processInput(
								"Error executing command, see previous output for details. Please, try again.",
							)
							continue
						}
					} else {
						lastResult = nil
					}
				}

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
}
