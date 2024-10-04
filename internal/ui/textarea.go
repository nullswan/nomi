package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
)

// NewTextArea initializes a new text area component.
func NewTextArea() textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "Type here..."
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.MaxWidth = 0

	ta.Focus()

	return ta
}
