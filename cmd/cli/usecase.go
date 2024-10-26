package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/cli"
	"github.com/nullswan/nomi/internal/logger"
	"github.com/nullswan/nomi/internal/tools"
	"github.com/nullswan/nomi/usecases/commit"
	"github.com/nullswan/nomi/usecases/copywriter"
	"github.com/spf13/cobra"
)

var usecaseCmd = &cobra.Command{
	Use:     "usecase [usecaseID]",
	Aliases: []string{"u"},
	Short:   "Run a usecase",
	Run: func(_ *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("You must provide a usecase")
			return
		}

		usecaseID := args[0]

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			fmt.Println("Sig received, quitting...")
			cancel()
		}()

		console := tools.NewBashConsole()
		selector := tools.NewSelector()
		toolsLogger := tools.NewLogger(
			cfg.DevMode,
		)

		// Initialize Providers
		logger := logger.Init()

		chatRepo, err := cli.InitChatDatabase(cfg.Output.Sqlite.Path)
		if err != nil {
			logger.With("error", err).
				Error("Error creating chat repository")
			os.Exit(1)
		}
		defer chatRepo.Close()

		conversation := chat.NewStackedConversation(chatRepo)
		inputArea := tools.NewInputArea()

		textToJSONBackend, err := cli.InitJSONProviders(
			logger,
			targetModel,
		)
		if err != nil {
			fmt.Printf("Error initializing providers: %v\n", err)
			return
		}
		defer textToJSONBackend.Close()

		ttjBackend := tools.NewTextToJSONBackend(
			textToJSONBackend,
			logger,
		)

		switch usecaseID {
		case "commit":
			err = commit.OnStart(
				ctx,
				console,
				selector,
				toolsLogger,
				ttjBackend,
				inputArea,
				conversation,
			)
		case "copywriter":
			err = copywriter.OnStart(
				ctx,
				selector,
				toolsLogger,
				inputArea,
				ttjBackend,
				conversation,
			)
		default:
			fmt.Println("usecase not found")
			return
		}

		if err != nil {
			fmt.Println("Error running usecase:", err)
		}
	},
}

var usecaseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all usecases",
	Run: func(_ *cobra.Command, _ []string) {
	},
}

var usecaseAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new usecase",
	Run:   func(_ *cobra.Command, _ []string) {},
}
