package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var conversationCmd = &cobra.Command{
	Use:   "conversation",
	Short: "Manage conversations",
}

var conversationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all conversations",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing conversations...")
		// TODO: Implement listing conversations
	},
}
