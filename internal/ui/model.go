package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	stopwatchIntval = time.Millisecond * 100
	loadingMessage  = "Loading..."
	aiPrefix        = "AI: "
	humanPrefix     = "You: "
)

// Define the model struct
type model struct {
	// textInput is the text input component.
	// It is used to capture user input.
	textArea textarea.Model

	// Pager is the pager component.
	// The pager is used to display the text received.
	pagerContent          string
	pager                 viewport.Model
	pagerRenderer         *glamour.TermRenderer
	ready                 bool
	pagerStopwatch        stopwatch.Model
	pagerAIBuffer         string
	pagerAIRenderedBuffer string

	providerName   string
	lastAiResponse string
	lastKey        string

	// commandCh is the channel where the user input is sent.
	commandCh chan string

	humanStyle lipgloss.Style
	humanText  lipgloss.Style
	aiStyle    lipgloss.Style
}

// NewModel initializes a new model with the provided channels.
func NewModel(inputChan chan string, providerName string) model {
	return model{
		textArea:              NewTextArea(),
		pagerContent:          "",
		pager:                 viewport.Model{}, // This is a dummy pager, it will be initialized in the Update function
		pagerRenderer:         nil,              // We initialize this in the Init function
		ready:                 false,
		pagerStopwatch:        stopwatch.NewWithInterval(stopwatchIntval),
		pagerAIBuffer:         "",
		pagerAIRenderedBuffer: "",
		providerName:          providerName,
		lastAiResponse:        "",
		lastKey:               "",
		commandCh:             inputChan,
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
		tea.SetWindowTitle("Golem 🗿"),
		func() tea.Msg {
			handlePipedInput(m.commandCh)
			return nil
		},
	)
}
