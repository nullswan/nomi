package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

// Define the model struct
type model struct {
	// textInput is the text input component.
	// It is used to capture user input.
	textArea textarea.Model

	// Pager is the pager component.
	// The pager is used to display the text received.
	pagerContent string
	pager        viewport.Model
	ready        bool

	// commandChannel is the channel where the user input is sent.
	commandChannel chan string

	humanStyle lipgloss.Style
	humanText  lipgloss.Style
	aiStyle    lipgloss.Style
}

// NewModel initializes a new model with the provided channels.
func NewModel(inputChan chan string) model {
	return model{
		textArea:       NewTextArea(),
		pagerContent:   "",
		pager:          viewport.Model{},
		ready:          false,
		commandChannel: inputChan,
		humanStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF00FF")),
		humanText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")),
		aiStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")),
	}
}

// Init initializes the model and returns any initial commands.
func (m model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
	)
}
