package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/jedib0t/go-pretty/v6/table"
	"gopkg.in/yaml.v2"

	prompts "github.com/nullswan/golem/internal/prompt"
	"github.com/spf13/cobra"
)

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Manage prompts",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("Use 'golem prompt list' to list available prompts.")
	},
}

var promptAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new prompt",
	Long:  `Add a new prompt to the system.`,
	Run: func(_ *cobra.Command, args []string) {
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
	Run: func(_ *cobra.Command, args []string) {
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

var promptEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Edit an existing prompt",
	Long:  `Edit an existing prompt by its ID.`,
	Run: func(_ *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please provide the ID of the prompt to edit.")
			return
		}
		id := args[0]

		// Fetch the current prompt by ID
		existingPrompt, err := prompts.LoadPrompt(id)
		if err != nil {
			fmt.Printf("Error fetching prompt: %v\n", err)
			return
		}

		// Write the current prompt to a temporary YAML file
		tempFile, err := os.CreateTemp("/tmp", "*.yaml")
		if err != nil {
			fmt.Printf("Error creating temporary file: %v\n", err)
			return
		}
		defer os.Remove(tempFile.Name()) // Clean up the file afterwards

		promptYaml, err := yaml.Marshal(existingPrompt)
		if err != nil {
			fmt.Printf("Error marshalling prompt to YAML: %v\n", err)
			return
		}

		// Write the YAML to the temp file
		_, err = tempFile.Write(promptYaml)
		if err != nil {
			fmt.Printf("Error writing to temp file: %v\n", err)
			return
		}

		// Close the file to ensure all data is flushed
		tempFile.Close()

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		process := exec.Command(editor, tempFile.Name())
		process.Stdin = os.Stdin
		process.Stdout = os.Stdout
		process.Stderr = os.Stderr

		// Open the temp file in Vim
		if err := process.Run(); err != nil {
			fmt.Printf("Error opening Vim: %v\n", err)
			return
		}

		// Read the updated content
		updatedData, err := os.ReadFile(tempFile.Name())
		if err != nil {
			fmt.Printf("Error reading updated file: %v\n", err)
			return
		}

		// Unmarshal the updated YAML back to the prompt struct
		var updatedPrompt prompts.Prompt
		if err := yaml.Unmarshal(updatedData, &updatedPrompt); err != nil {
			fmt.Printf("Error unmarshalling updated YAML: %v\n", err)
			return
		}

		// Validate the updated prompt
		if err := updatedPrompt.Validate(); err != nil {
			fmt.Printf("Validation error: %v\n", err)
			return
		}

		// Save the updated prompt
		if err := updatedPrompt.Save(); err != nil {
			fmt.Printf("Error saving updated prompt: %v\n", err)
			return
		}

		fmt.Println("Prompt edited successfully.")

		promptYaml, err = yaml.Marshal(updatedPrompt)
		if err != nil {
			fmt.Printf("Error marshalling prompt to YAML: %v\n", err)
			return
		}

		fmt.Println("---\n" + string(promptYaml))
	},
}
