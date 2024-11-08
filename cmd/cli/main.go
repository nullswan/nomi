package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
			return
		}
	}

	// Initialize Providers
	textToTextBackend, err := cli.InitTextProviders(
		logger,
		targetModel,
		selectedPrompt.Preferences.Reasoning,
	)
	if err != nil {
		fmt.Printf("Error initializing providers: %v\n", err)
		return
	}
	defer textToTextBackend.Close()

	// Initialize Database
	repo, err := cli.InitChatDatabase(
		cfg.Output.Sqlite.Path,
	)
	if err != nil {
		fmt.Printf("Error creating repository: %v\n", err)
		return
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
		return
	}

	// Prepare the welcome message
	welcomeConfig := cli.NewWelcomeConfig(
		conversation,
		cli.WithWelcomeMessage("Nomi is ready to assist you."),
		cli.WithBuildDate(buildDate),
		cli.WithBuildVersion(buildVersion),
		cli.WithStartPrompt(startPrompt),
		cli.WithModelProvider(textToTextBackend),
		cli.WithProvider(providers.CheckProvider()),
	)

	// Initialize Renderer
	renderer, err := term.InitRenderer()
	if err != nil {
		fmt.Printf("Error initializing renderer: %v\n", err)
		return
	}

	inputCh := make(chan string)
	inputErrCh := make(chan error)
	readyCh := make(chan struct{})
	voiceInputCh := make(chan string)

	var inputStream *audio.StreamHandler
	var audioStartCh, audioEndCh <-chan struct{}
	if cfg.Input.Voice.Enabled {
		// Initialize Voice using shared method
		inputStream, audioStartCh, audioEndCh, err = cli.InitVoice(
			cfg,
			logger,
			func(text string, isProcessing bool) {
				if !isProcessing {
					fmt.Println(">>>", text)
					voiceInputCh <- text
				}
			},
			cfg.Input.Voice.KeyCode,
			cfg.Input.Voice.Language,
		)
		if err != nil {
			fmt.Printf("Error initializing voice: %v\n", err)
			return
		}
		defer inputStream.Close()

		cli.WithAdditionalLine(
			"  Voice language: " + cfg.Input.Voice.Language,
		)(
			&welcomeConfig,
		)
	}

	// Display Welcome Message
	if !interactiveMode {
		cli.DisplayWelcome(welcomeConfig)
	}

	// Start Input Reader Goroutine
	go term.ReadInput(inputCh, inputErrCh, readyCh)

	// Main Event Loop
	cli.EventLoop(
		ctx,
		cancel,
		inputCh,
		inputErrCh,
		readyCh,
		voiceInputCh,
		audioStartCh,
		audioEndCh,
		inputStream,
		logger,
		conversation,
		renderer,
		textToTextBackend,
		processInput,
	)
}

func processInput(
	ctx context.Context,
	text string,
	conversation chat.Conversation,
	renderer *term.Renderer,
	textToTextBackend baseprovider.TextToTextProvider,
) {
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
