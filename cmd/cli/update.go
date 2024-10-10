package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Golem to the latest version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("Updating Golem...")
	},
}
