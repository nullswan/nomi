package main

import (
	"context"
	"fmt"
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
	// TODO(nullswan): Should be configurable
	cmdKeyCode = 55
)

var rootCmd = &cobra.Command{
	Use:   "nomi [flags] [arguments]",
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
		cli.DisplayWelcome(cli.NewWelcomeConfig(
			conversation,
			cli.WithWelcomeMessage("Nomi is ready to assist you."),
			cli.WithBuildDate(buildDate),
			cli.WithBuildVersion(buildVersion),
			cli.WithStartPrompt(startPrompt),
			cli.WithModelProvider(textToTextBackend),
			cli.WithProvider(providers.CheckProvider()),
			cli.WithDefaultIntrustructions(),
		))
	}

	// Initialize Renderer
	renderer, err := term.InitRenderer()
	if err != nil {
		fmt.Printf("Error initializing renderer: %v\n", err)
		os.Exit(1)
	}

	// Initialize Audio
	// TODO(nullswan): Only when using audio, till local whisper is supported
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
	go cli.ReadInput(rl, inputCh, inputErrCh)

	// Initialize Key Hooks
	audioStartCh, audioEndCh := cli.SetupKeyHooks(cmdKeyCode)

	// Main Event Loop
	cli.EventLoop(
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
		processInput,
	)
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
