package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/chzyer/readline"
	"github.com/gordonklaus/portaudio"
	"github.com/nullswan/nomi/internal/audio"
	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/cli"
	"github.com/nullswan/nomi/internal/code"
	"github.com/nullswan/nomi/internal/completion"
	"github.com/nullswan/nomi/internal/config"
	"github.com/nullswan/nomi/internal/logger"
	"github.com/nullswan/nomi/internal/providers"
	baseprovider "github.com/nullswan/nomi/internal/providers/base"
	"github.com/nullswan/nomi/internal/term"
	"github.com/spf13/cobra"
)

const (
	interpreterMaxRetries = 3
)

var interpreterCmd = &cobra.Command{
	Use:   "interpreter",
	Short: "Start the interpreter",
	Run: func(_ *cobra.Command, _ []string) {
		log := logger.Init()

		// Ensure the provider is OpenAI
		provider := providers.CheckProvider()
		if provider != providers.OpenAIProvider {
			log.Error("Error: only openai is supported currently.")
			os.Exit(1)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			log.With("error", err).Error("Error loading configuration")
			os.Exit(1)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		interpreterAskPrompt, err := code.GetDefaultInterpreterPrompt(
			runtime.GOOS,
		)
		if err != nil {
			log.With("error", err).
				Error("Error getting default interpreter prompt")
			os.Exit(1)
		}

		codeGenerationBackend, err := cli.InitProviders(
			log,
			"",
			interpreterAskPrompt.Preferences.Reasoning,
		)
		if err != nil {
			log.With("error", err).
				Error("Error initializing code generation providers")
			os.Exit(1)
		}

		codeInferenceBackend, err := cli.InitProviders(
			log,
			"",
			false,
		)
		if err != nil {
			log.With("error", err).
				Error("Error initializing code inference providers")
			os.Exit(1)
		}

		chatRepo, err := cli.InitChatDatabase(cfg.Output.Sqlite.Path)
		if err != nil {
			log.With("error", err).
				Error("Error creating chat repository")
			os.Exit(1)
		}
		defer chatRepo.Close()

		codeRepo, err := cli.InitCodeDatabase(cfg.Output.Sqlite.Path)
		if err != nil {
			log.With("error", err).
				Error("Error creating code repository")

			os.Exit(1)
		}
		defer codeRepo.Close()

		conversation, err := cli.InitConversation(
			chatRepo,
			nil,
			interpreterAskPrompt,
		)
		if err != nil {
			fmt.Printf("Error initializing conversation: %v\n", err)
			os.Exit(1)
		}

		welcomeConfig := cli.NewWelcomeConfig(
			conversation,
			cli.WithBuildDate(buildDate),
			cli.WithBuildVersion(buildVersion),
			cli.WithStartPrompt(startPrompt),
			cli.WithModelProvider(codeGenerationBackend),
			cli.WithModelProvider(codeInferenceBackend),
			cli.WithProvider(provider),
			cli.WithDefaultIntrustructions(),
		)

		// Initialize Readline
		rl, err := term.InitReadline()
		if err != nil {
			fmt.Printf("Error initializing readline: %v\n", err)
			os.Exit(1)
		}
		defer rl.Close()

		inputCh := make(chan string)
		inputErrCh := make(chan error)

		var inputStream *audio.AudioStream
		var audioStartCh, audioEndCh <-chan struct{}

		if cfg.Input.Voice.Enabled {
			// Initialize Voice using shared method
			inputStream, audioStartCh, audioEndCh, err = cli.InitVoice(
				cfg,
				log,
				func(text string, isProcessing bool) {
					rl.Operation.Clean()
					if !isProcessing {
						rl.Operation.SetBuffer("")
						fmt.Printf("%s\n\n", text)
						inputCh <- text
					} else {
						rl.Operation.SetBuffer(text)
					}
				},
				cmdKeyCode,
			)
			if err != nil {
				fmt.Printf("Error initializing voice: %v\n", err)
				os.Exit(1)
			}
			defer inputStream.Close()

			if inputStream != nil {
				defer portaudio.Terminate()
			}

			cli.WithVoiceInstructions()(&welcomeConfig)
		}

		oaiKey := os.Getenv("OPENAI_API_KEY")
		if oaiKey == "" {
			fmt.Println("OPENAI_API_KEY is not set")
			os.Exit(1)
		}

		renderer, err := term.InitRenderer()
		if err != nil {
			log.With("error", err).
				Error("Error initializing renderer")
			os.Exit(1)
		}

		availableBlocks, err := codeRepo.LoadCodeBlocks()
		if err != nil {
			log.With("error", err).
				Error("Error loading code blocks")
			return
		}

		blockMap := make(map[string]code.CodeBlock, len(availableBlocks))
		for _, block := range availableBlocks {
			blockMap[block.ID] = block
		}

		if !interactiveMode {
			cli.DisplayWelcome(welcomeConfig)
		}

		retries := 0
		processInput := func(
			ctx context.Context,
			text string,
			conv chat.Conversation,
			renderer *term.Renderer,
			_ baseprovider.TextToTextProvider,
			rl *readline.Instance,
		) {
			defer rl.Refresh()

			var lastResult []code.ExecutionResult
			text = cli.HandleCommands(text, conv)
			if text == "" {
				return
			}

			conv.AddMessage(chat.NewMessage(chat.RoleUser, text))

			if len(blockMap) > 0 && retries == 0 {
				log.Debug("Trying to get suggestion from available code blocks")

				var cachedBlock *code.CodeBlock
				block, err := getSuggestionFromBlocks(
					text,
					codeInferenceBackend,
					availableBlocks,
				)
				if err != nil {
					log.With("error", err).
						Error("Error getting suggestion from blocks")
				} else if block != nil {
					cachedBlock = block
				}

				if cachedBlock != nil {
					execResult := code.ExecuteCodeBlock(*cachedBlock)
					lastResult = []code.ExecutionResult{execResult}
					fmt.Printf(
						"Received (%d): %s\n%s\n",
						execResult.ExitCode,
						execResult.Stdout,
						execResult.Stderr,
					)
					return
				}
			}

			fmt.Printf(
				"I don't know the answer to that question. Let me try to find out...\n",
			)

			completion, err := cli.GenerateCompletion(
				ctx,
				conv,
				renderer,
				codeGenerationBackend,
			)
			if err != nil {
				if strings.Contains(err.Error(), "context canceled") {
					fmt.Println("\nRequest canceled by the user.")
					return
				}
				log.With("error", err).
					Error("Error generating completion")

				return
			}

			conv.AddMessage(
				chat.NewMessage(chat.RoleAssistant, completion),
			)

			// Exec the command
			result := code.InterpretCodeBlocks(completion)
			for _, r := range result {
				if r.ExitCode == 0 && r.Stderr == "" {
					description, err := storeCodePrompt(
						codeInferenceBackend,
						r.Block,
						codeRepo,
					)
					if err != nil {
						log.With("error", err).
							Error("Error storing code prompt")
						continue
					}

					r.Block.Description = description
					blockMap[r.Block.ID] = r.Block
					retries = 0
				} else {
					fmt.Printf("Error executing command: %s\n", r.Stderr)
				}

				// TODO(nullswan): Display in a better way
				fmt.Printf(
					"Received (%d): %s\n%s\n",
					r.ExitCode,
					r.Stdout,
					r.Stderr,
				)
			}

			if len(result) == 0 {
				fmt.Println("No results received.")
				return
			}

			// Extend the conversation with the result
			formattedResult := code.FormatExecutionResultForLLM(result)
			conv.AddMessage(
				chat.NewMessage(chat.RoleAssistant, formattedResult),
			)

			lastResult = result

			// Handle retries by pushing back to inputCh
			containsError := false
			for _, res := range lastResult {
				if res.ExitCode != 0 {
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
				} else {
					fmt.Printf("Retrying...\n")
					inputCh <- "Error executing command, see previous output for details. Please, try again."
				}
			}
		}

		go cli.ReadInput(rl, inputCh, inputErrCh)

		cli.EventLoop(
			ctx,
			cancel,
			inputCh,
			inputErrCh,
			audioStartCh,
			audioEndCh,
			inputStream,
			log,
			conversation,
			renderer,
			codeGenerationBackend,
			rl,
			processInput,
		)
	},
}

func storeCodePrompt(
	textToTextBackend baseprovider.TextToTextProvider,
	block code.CodeBlock,
	repo code.Repository,
) (string, error) {
	outCh := make(chan completion.Completion)
	messages := []chat.Message{
		{
			Role:    chat.RoleSystem,
			Content: code.DefaultInterpreterInferencePrompt.Settings.SystemPrompt,
		},
		{
			Role:    chat.RoleUser,
			Content: block.Code,
		},
	}

	go func() {
		defer close(outCh)
		if err := textToTextBackend.GenerateCompletion(
			context.Background(),
			messages,
			outCh,
		); err != nil {
			if strings.Contains(err.Error(), "context canceled") {
				return
			}
			fmt.Printf("Error generating completion: %v\n", err)
		}
	}()

	for {
		select {
		case cmpl, ok := <-outCh:
			if !ok {
				return "", fmt.Errorf("error generating completion")
			}

			if !completion.IsTombStone(cmpl) {
				continue
			}

			block.Description = cmpl.Content()
			err := repo.SaveCodeBlock(block)
			if err != nil {
				fmt.Printf("Error saving code block: %v\n", err)
			}

			return block.Description, nil
		}
	}
}

func getSuggestionFromBlocks(
	prompt string,
	inferenceProvider baseprovider.TextToTextProvider,
	availableBlocks []code.CodeBlock,
) (*code.CodeBlock, error) {
	outCh := make(chan completion.Completion)

	availableBlocksText := ""
	for _, block := range availableBlocks {
		availableBlocksText += fmt.Sprintf(
			"%s: %s\n",
			block.ID,
			block.Description,
		)
	}

	messages := []chat.Message{
		{
			Role:    chat.RoleSystem,
			Content: code.DefaultInterpreterCachePrompt.Settings.SystemPrompt + "\n----\nAvailable blocks:\n" + availableBlocksText,
		},
		{
			Role:    chat.RoleUser,
			Content: "Question:\n" + prompt,
		},
	}

	go func() {
		defer close(outCh)
		if err := inferenceProvider.GenerateCompletion(
			context.Background(),
			messages,
			outCh,
		); err != nil {
			if strings.Contains(err.Error(), "context canceled") {
				return
			}

			fmt.Printf("Error generating completion: %v\n", err)
		}
	}()

	for {
		select {
		case cmpl, ok := <-outCh:
			if !ok {
				return nil, fmt.Errorf("error generating completion")
			}

			if !completion.IsTombStone(cmpl) {
				continue
			}

			fmt.Printf("Sent: %s\n", prompt)
			fmt.Printf("Received: %s\n", cmpl.Content())
			for _, block := range availableBlocks {
				if block.ID == cmpl.Content() {
					return &block, nil
				}
			}

			return nil, nil
		}
	}
}
