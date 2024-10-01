package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var presetCmd = &cobra.Command{
	Use:   "preset",
	Short: "Manage presets",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use 'yourapp preset list' to list available presets.")
	},
}

var presetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all presets",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing presets...")
		// TODO: Implement listing presets
	},
}
