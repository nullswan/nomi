package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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

			var err error
			m.pagerRenderer, err = glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(int(msg.Width)),
			)
			if err != nil {
				return m, nil
			}

			m.textArea.SetWidth(msg.Width)

			m.ready = true
		} else {
			m.pager.Width = msg.Width
			m.pager.Height = msg.Height - verticalMarginHeight
		}
	case PagerMsg:
		// Add the message to the pager content
		if msg.From == Human {
			str, err := m.pagerRenderer.Render(msg.String())
			if err != nil {
				return m, nil
			}

			m.pagerContent += m.humanStyle.Render(humanPrefix) + "\n" + str + "\n"
		} else if msg.From == AI {
			if msg.String() == "" {
				// The AI stopped talking, so stop the stopwatch
				cmds = append(cmds, m.pagerStopwatch.Stop())
				m.pagerAIRenderedBuffer = ""
				m.pagerAIBuffer = ""
			} else {
				// The AI is talking
				m.pagerAIBuffer += msg.String()

				if m.pagerAIRenderedBuffer == "" {
					m.pagerContent += m.aiStyle.Render(aiPrefix) + "\n"
				} else {
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
		m.pager.GotoBottom() // TODO(nullswan): Don't go to the bottom if the user has scrolled up

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
