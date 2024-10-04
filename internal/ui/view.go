package ui

import "fmt"

func (m model) View() string {
	if !m.ready {
		return "Loading..."
	}

	return fmt.Sprintf(
		"%s\n%s\n%s",
		m.headerView(),
		m.pager.View(),
		m.footerView(),
	) + "\n\nInput (type or use voice):\n" + m.textArea.View()
}
