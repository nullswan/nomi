package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.textArea.Focused() {
				m.textArea.Blur()
			}

		case tea.KeyEnter:
			// TODO(nullswan): Handle KeyEnter+Shift, KeyEnter+Ctrl, etc.
			if m.textArea.Focused() {
				m.textArea.Blur()
			}

			// If the text area is empty, do nothing
			if m.textArea.Value() == "" {
				return m, nil
			}

			// Send the command
			m.commandChannel <- m.textArea.Value()
			m.textArea.Reset()

			cmds = append(cmds, m.pagerStopwatch.Reset())
			cmds = append(cmds, m.pagerStopwatch.Start())
		case tea.KeyCtrlC:
			return m, tea.Quit
		default:
			if !m.textArea.Focused() {
				cmd = m.textArea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight + 10

		if !m.ready {
			m.pager = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.pager.YPosition = headerHeight
			m.pager.SetContent(m.pagerContent)

			m.textArea.SetWidth(msg.Width)

			// Override the default glamour style with a dark one
			// TODO(nullswan): Handle light theme
			var styleOpt glamour.TermRendererOption
			darkStyle := styles.DarkStyleConfig
			darkStyle.Document.Margin = uintPtr(0)
			styleOpt = glamour.WithStyles(darkStyle)

			// Initialize the glamour renderer
			var err error
			m.pagerRenderer, err = glamour.NewTermRenderer(
				styleOpt,
				glamour.WithWordWrap(0),
				glamour.WithEmoji(),
			)
			if err != nil {
				fmt.Printf("Error initializing renderer: %v\n", err)
			}

			m.ready = true
		}

		m.pager.Width = msg.Width
		m.pager.Height = msg.Height - verticalMarginHeight
		m.textArea.SetWidth(msg.Width)
	case PagerMsg:
		isAtBottom := m.pager.AtBottom()

		if msg.From == Human {
			str, err := m.pagerRenderer.Render(msg.String())
			if err != nil {
				return m, nil
			}

			m.pagerContent += m.humanStyle.Render(humanPrefix) + "\n" + str + "\n"
		} else if msg.From == AI {
			if msg.Stop() {
				// The AI stopped talking, so stop the stopwatch
				cmds = append(cmds, m.pagerStopwatch.Stop())
				m.pagerAIRenderedBuffer = ""
				m.pagerAIBuffer = ""
			} else {
				// The AI is talking
				m.pagerAIBuffer += msg.String()

				// If the AI is starting to talk, render the AI style
				if m.pagerAIRenderedBuffer == "" {
					m.pagerContent += m.aiStyle.Render(aiPrefix) + "\n"
				} else {
					// if human prefix is contained in the content that is going to be removed
					// it means that the context has been canceled
					bufferToRemove := m.pagerContent[len(m.pagerContent)-len(m.pagerAIRenderedBuffer):]
					if strings.Contains(bufferToRemove, humanPrefix) {
						cmds = append(cmds, m.pagerStopwatch.Stop())
						m.pagerAIRenderedBuffer = ""
						m.pagerAIBuffer = ""
						return m, nil
					}

					// Remove the old AI content, to replace it with the new one
					m.pagerContent = m.pagerContent[:len(m.pagerContent)-len(m.pagerAIRenderedBuffer)]
				}

				newContent, err := m.pagerRenderer.Render(m.pagerAIBuffer)
				if err != nil {
					return m, nil
				}

				m.pagerAIRenderedBuffer = newContent
				m.pagerContent += newContent
			}
		}

		// Update the pager content
		m.pager.SetContent(m.pagerContent)
		if isAtBottom || msg.From == Human {
			m.pager.GotoBottom()
		}
		return m, tea.Batch(cmds...)
	}

	m.pagerStopwatch, cmd = m.pagerStopwatch.Update(msg)
	cmds = append(cmds, cmd)

	// Update the text area
	m.textArea, cmd = m.textArea.Update(msg)
	cmds = append(cmds, cmd)

	// Update the pager
	m.pager, cmd = m.pager.Update(msg)
	cmds = append(cmds, cmd)

	// Return the updated model and any commands to run
	return m, tea.Batch(cmds...)
}

func uintPtr(u uint) *uint { return &u }
