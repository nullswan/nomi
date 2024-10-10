package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/manifoldco/promptui/screenbuf"
	"github.com/nullswan/golem/internal/chat"
	"github.com/nullswan/golem/internal/term"
	"github.com/spf13/cobra"
)

var conversationCmd = &cobra.Command{
	Use:   "conversation",
	Short: "Manage conversations",
	Run: func(cmd *cobra.Command, _ []string) {
		err := cmd.Help()
		if err != nil {
			fmt.Println("Error displaying help:", err)
		}
	},
}

var conversationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all conversations",
	Long:  `List all available conversations with their ID and Created At.`,
	Run: func(_ *cobra.Command, _ []string) {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleLight)

		t.Style().Options.SeparateHeader = false
		t.Style().Options.SeparateFooter = false
		t.Style().Options.DrawBorder = false
		t.Style().Options.SeparateRows = false
		t.Style().Options.SeparateColumns = true
		t.Style().Options.SeparateColumns = false

		t.AppendHeader(
			table.Row{"Id", "Created At", "Since", "Messages"},
		)

		repo, err := chat.NewSQLiteRepository(cfg.Output.Sqlite.Path)
		if err != nil {
			fmt.Println("Error creating repository:", err)
			return
		}
		defer repo.Close()

		allConversations, err := repo.GetConversations()
		if err != nil {
			fmt.Println("Error listing conversations:", err)
			return
		}

		for _, convo := range allConversations {
			t.AppendRow(
				[]interface{}{
					convo.GetID(),
					convo.GetCreatedAt().Format(time.RFC3339),
					time.Since(convo.GetCreatedAt()).Round(time.Second),
					len(convo.GetMessages()),
				},
			)
		}

		t.Render()
	},
}

var conversationDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a conversation",
	Long:  `Delete a conversation by its ID.`,
	Run: func(_ *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please provide the ID of the conversation to delete.")
			return
		}
		id := args[0]

		repo, err := chat.NewSQLiteRepository(cfg.Output.Sqlite.Path)
		if err != nil {
			fmt.Println("Error creating repository:", err)
			return
		}
		defer repo.Close()

		err = repo.DeleteConversation(id)
		if err != nil {
			fmt.Println("Error deleting conversation:", err)
			return
		}

		fmt.Println("Conversation deleted.")
	},
}

var conversationShowCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show a conversation",
	Long:  `Show a conversation by its ID.`,
	Run: func(_ *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please provide the ID of the conversation to show.")
			return
		}
		id := args[0]

		repo, err := chat.NewSQLiteRepository(cfg.Output.Sqlite.Path)
		if err != nil {
			fmt.Println("Error creating repository:", err)
			return
		}
		defer repo.Close()

		convo, err := repo.LoadConversation(id)
		if err != nil {
			fmt.Println("Error showing conversation:", err)
			return
		}

		renderer, err := term.InitRenderer()
		if err != nil {
			fmt.Println("Error initializing renderer:", err)
			return
		}

		for _, msg := range convo.GetMessages() {
			fmt.Printf("%s:\n", msg.Role.String())

			sb := screenbuf.New(os.Stdout)

			lines := strings.Split(msg.Content, "\n")
			for _, line := range lines {
				sb.WriteString(line)
			}

			sb.Clear()

			mdContent, err := renderer.Render(msg.Content)
			if err != nil {
				fmt.Println("Error rendering markdown:", err)
				return
			}
			fmt.Println(mdContent)
		}
	},
}
