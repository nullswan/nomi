package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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

			if m.textArea.Value() == "" {
				return m, nil
			}

			m.commandChannel <- m.textArea.Value()
			m.textArea.Reset()
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
			m.ready = true
		} else {
			m.pager.Width = msg.Width
			m.pager.Height = msg.Height - verticalMarginHeight
		}

	case PagerMsg:
		str, err := m.pagerRenderer.Render(msg.String())
		if err != nil {
			return m, nil
		}

		// Add the message to the pager content
		if msg.From == Human {
			m.pagerContent += m.humanStyle.Render("You:") + "\n" + str + "\n"
		} else {
			m.pagerContent += m.aiStyle.Render("AI:") + "\n" + str + "\n"
		}

		// Update the pager content
		m.pager.SetContent(m.pagerContent)
		return m, nil
	}

	// Update the text area
	m.textArea, cmd = m.textArea.Update(msg)
	cmds = append(cmds, cmd)

	// Update the pager
	m.pager, cmd = m.pager.Update(msg)
	cmds = append(cmds, cmd)

	// Return the updated model and any commands to run
	return m, tea.Batch(cmds...)
}
