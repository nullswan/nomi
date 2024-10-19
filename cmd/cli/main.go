package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "net/http/pprof"

	"github.com/chzyer/readline"
	"github.com/gordonklaus/portaudio"
	"github.com/nullswan/nomi/internal/audio"
	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/cli"
	"github.com/nullswan/nomi/internal/config"
	"github.com/nullswan/nomi/internal/logger"
	prompts "github.com/nullswan/nomi/internal/prompt"
	"github.com/nullswan/nomi/internal/providers"
	baseprovider "github.com/nullswan/nomi/internal/providers/base"
	"github.com/nullswan/nomi/internal/term"

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
	binName = "nomi"
	// TODO(nullswan): Should be configurable
	cmdKeyCode = 55
)

var rootCmd = &cobra.Command{
	Use:   binName + " [flags] [arguments]",
	Short: "AI runtime, multi-modal, supporting action & private data. ",
	Run:   runApp,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func runApp(_ *cobra.Command, _ []string) {
	// Setup context and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Sig received, quitting...")
		cancel()
	}()

	// Initialize Logger
	logger := logger.Init()

	selectedPrompt := &prompts.DefaultPrompt
	if startPrompt != "" {
		var err error
		selectedPrompt, err = prompts.LoadPrompt(startPrompt)
		if err != nil {
			fmt.Printf("Error loading prompt: %v\n", err)
			os.Exit(1)
		}
	}

	// Initialize Providers
	textToTextBackend, err := cli.InitProviders(
		logger,
		targetModel,
		selectedPrompt.Preferences.Reasoning,
	)
	if err != nil {
		fmt.Printf("Error initializing providers: %v\n", err)
		os.Exit(1)
	}
	defer textToTextBackend.Close()

	// Initialize Database
	repo, err := cli.InitChatDatabase(
		cfg.Output.Sqlite.Path,
	)
	if err != nil {
		fmt.Printf("Error creating repository: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Initialize Repository and Conversation
	conversation, err := cli.InitConversation(
		repo,
		&startConversationID,
		*selectedPrompt,
	)
	if err != nil {
		fmt.Printf("Error initializing conversation: %v\n", err)
		os.Exit(1)
	}

	// Display Welcome Message
	if !interactiveMode {
		displayWelcome(conversation, textToTextBackend.GetModel())
	}

	// Initialize Renderer
	renderer, err := term.InitRenderer()
	if err != nil {
		fmt.Printf("Error initializing renderer: %v\n", err)
		os.Exit(1)
	}

	// Initialize Audio
	oaiKey := os.Getenv("OPENAI_API_KEY")
	if oaiKey == "" {
		fmt.Println("OPENAI_API_KEY is not set")
		os.Exit(1)
	}

	if err := portaudio.Initialize(); err != nil {
		fmt.Printf("Failed to initialize PortAudio: %v\n", err)
		os.Exit(1)
	}
	defer portaudio.Terminate()

	audioOpts, err := audio.ComputeAudioOptions(&audio.AudioOptions{})
	if err != nil {
		fmt.Printf("Error computing audio options: %v\n", err)
		os.Exit(1)
	}

	// Initialize Readline
	rl, err := term.InitReadline()
	if err != nil {
		fmt.Printf("Error initializing readline: %v\n", err)
		os.Exit(1)
	}
	defer rl.Close()

	inputCh := make(chan string)
	inputErrCh := make(chan error)

	// Initialize Transcription Server
	ts, err := cli.InitTranscriptionServer(
		oaiKey,
		audioOpts,
		logger,
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
	)
	if err != nil {
		fmt.Printf("Error initializing transcription server: %v\n", err)
		os.Exit(1)
	}
	defer ts.Close()
	ts.Start()

	// Initialize VAD
	vad := cli.InitVAD(ts, logger)
	defer vad.Stop()
	vad.Start()

	// Create Input Stream
	inputStream, err := audio.NewInputStream(
		logger,
		audioOpts,
		func(buffer []float32) {
			vad.Feed(buffer)
		},
	)
	if err != nil {
		fmt.Printf("Failed to create input stream: %v\n", err)
		os.Exit(1)
	}

	defer inputStream.Close()

	// Start Input Reader Goroutine
	go readInput(rl, inputCh, inputErrCh)

	// Initialize Key Hooks
	audioStartCh, audioEndCh := cli.SetupKeyHooks(cmdKeyCode)

	// Main Event Loop
	eventLoop(
		ctx,
		cancel,
		inputCh,
		inputErrCh,
		audioStartCh,
		audioEndCh,
		inputStream,
		logger,
		conversation,
		renderer,
		textToTextBackend,
		rl,
	)
}

func displayWelcome(conversation chat.Conversation, model string) {
	fmt.Printf("----\n")
	fmt.Printf("Nomi (%s)\n", buildVersion)
	fmt.Println()
	fmt.Println("Configuration")
	fmt.Printf("  Start prompt: %s\n", startPrompt)
	fmt.Printf("  Conversation: %s\n", conversation.GetID())
	fmt.Printf("  Provider: %s\n", providers.CheckProvider())
	fmt.Printf("  Model: %s\n", model)
	fmt.Printf("  Build Date: %s\n", buildDate)
	fmt.Printf("-----\n")
	fmt.Printf("Press Enter twice to send a message.\n")
	fmt.Printf("Press Ctrl+C to exit.\n")
	fmt.Printf("Press Ctrl+K to cancel the current request.\n")
	fmt.Printf("Press any[once] and CMD to record audio.\n")
	fmt.Printf("-----\n\n")
}

func readInput(
	rl *readline.Instance,
	inputCh chan<- string,
	inputErrCh chan<- error,
) {
	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				inputErrCh <- term.ErrInputInterrupted
				return
			}
			if err == io.EOF {
				// when killed, wait for alive..
				inputErrCh <- term.ErrInputKilled
				return
			}
			inputErrCh <- fmt.Errorf("error reading input: %w", err)
			return
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		inputCh <- line
	}
}

func eventLoop(
	ctx context.Context,
	cancel context.CancelFunc,
	inputCh chan string,
	inputErrCh chan error,
	audioStartCh, audioEndCh <-chan struct{},
	inputStream *audio.AudioStream,
	logger *logger.Logger,
	conversation chat.Conversation,
	renderer *term.Renderer,
	textToTextBackend baseprovider.TextToTextProvider,
	rl *readline.Instance,
) {
	audioRunning := false
	// var wg sync.WaitGroup

	defer func() {
		if audioRunning {
			inputStream.Stop()
		}
	}()

	eventCtx, eventCtxCancel := context.WithCancel(ctx)
	defer eventCtxCancel()

	for {
		select {
		case <-ctx.Done():
			return
		case line := <-inputCh:
			eventCtxCancel()
			eventCtx, eventCtxCancel = context.WithCancel(ctx)
			defer eventCtxCancel()

			processInput(
				eventCtx,
				line,
				conversation,
				renderer,
				textToTextBackend,
				rl,
			)
		case err := <-inputErrCh:
			if errors.Is(err, term.ErrInputInterrupted) ||
				errors.Is(err, term.ErrInputKilled) {
				cancel()
				break
			}
			fmt.Printf("Error reading input: %v\n", err)
		case <-audioStartCh:
			if !audioRunning {
				audioRunning = true
				// fmt.Println("Recording audio...")
				err := inputStream.Start()
				if err != nil {
					logger.
						With("error", err).
						Error("Failed to start input stream")
				}
				// closeReadline()
			}
		case <-audioEndCh:
			if audioRunning {
				audioRunning = false
				// fmt.Println("Audio recording stopped.")
				err := inputStream.Stop()
				if err != nil {
					logger.
						With("error", err).
						Error("Failed to stop input stream")
				}
				// reinitReadline(&wg, inputCh, inputErrCh)
			}
		}
	}
}

func setupReadline(
	rl *readline.Instance,
	inputCh chan<- string,
	inputErrCh chan<- error,
) {
	go readInput(rl, inputCh, inputErrCh)
}

func setupNewReadline() (*readline.Instance, error) {
	return term.InitReadline()
}

func processInput(
	ctx context.Context,
	text string,
	conversation chat.Conversation,
	renderer *term.Renderer,
	textToTextBackend baseprovider.TextToTextProvider,
	rl *readline.Instance,
) {
	defer rl.Refresh()

	if text == "" {
		return
	}

	text = cli.HandleCommands(text, conversation)
	if text == "" {
		return
	}

	conversation.AddMessage(chat.NewMessage(chat.RoleUser, text))

	completion, err := cli.GenerateCompletion(
		ctx,
		conversation,
		renderer,
		textToTextBackend,
	)
	if err != nil {
		if strings.Contains(err.Error(), "context canceled") {
			fmt.Println("\nRequest canceled by the user.")
			return
		}

		fmt.Printf("Error generating completion: %v\n", err)
		return
	}

	conversation.AddMessage(chat.NewMessage(chat.RoleAssistant, completion))
}
