package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"gopkg.in/yaml.v2"

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

var promptAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new prompt",
	Long:  `Add a new prompt to the system.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please provide the URL of the YAML file.")
			return
		}
		url := args[0]

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error fetching the URL: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error: received status code %d\n", resp.StatusCode)
			return
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response body: %v\n", err)
			return
		}

		var prompt prompts.Prompt
		err = yaml.Unmarshal(data, &prompt)
		if err != nil {
			fmt.Printf("Error unmarshalling YAML: %v\n", err)
			return
		}

		// Validate the prompt
		if err := prompt.Validate(); err != nil {
			fmt.Printf("Validation error: %v\n", err)
			return
		}

		// Save the prompt
		if err := prompt.Save(); err != nil {
			fmt.Printf("Error saving prompt: %v\n", err)
			return
		}

		fmt.Println("Prompt added successfully.")

		promptYaml, err := yaml.Marshal(prompt)
		if err != nil {
			fmt.Printf("Error marshalling prompt to YAML: %v\n", err)
			return
		}

		fmt.Println("---\n" + string(promptYaml))
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
			table.Row{"Id", "Name", "Description", "Version", "Author"},
		)

		allPrompts, err := prompts.ListPrompts()
		if err != nil {
			fmt.Println("Error listing prompts:", err)
			return
		}

		for _, prompt := range allPrompts {
			t.AppendRow(
				[]interface{}{
					prompt.ID,
					prompt.Name,
					prompt.Description,
					prompt.Metadata.Version,
					prompt.Metadata.Author,
				},
			)
		}

		t.Render()
	},
}
