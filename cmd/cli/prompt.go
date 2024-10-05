package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"

	prompts "github.com/nullswan/golem/internal/prompt"
	"github.com/spf13/cobra"
)

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Manage prompts",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use 'golem prompt list' to list available prompts.")
	},
}

var promptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all prompts",
	Long:  `List all available prompts with their ID, Name, and Description.`,
	Run: func(cmd *cobra.Command, args []string) {
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
			table.Row{"Id", "Name", "Description", "Version", "Updated"},
		)

		allPrompts, err := prompts.ListPrompts()
		if err != nil {
			fmt.Println("Error listing prompts:", err)
			return
		}

		for _, prompt := range allPrompts {
			since := time.Since(prompt.Metadata.UpdatedAt).Round(time.Second)

			t.AppendRow(
				[]interface{}{
					prompt.ID,
					prompt.Name,
					prompt.Description,
					prompt.Metadata.Version,
					since.String(),
				},
			)
		}

		t.Render()
	},
}
