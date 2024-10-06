package ui

import (
	"fmt"
	"time"
)

func (m model) View() string {
	if !m.ready {
		return loadingMessage
	}

	pager := fmt.Sprintf(
		"%s\n%s\n%s",
		m.headerView(),
		m.pager.View(),
		m.footerView(),
	)

	var textAreaHeader string
	if m.pagerStopwatch.Running() {
		elapsedTime := formatDuration(m.pagerStopwatch.Elapsed())
		renderedTime := m.aiStyle.Render(elapsedTime)
		textAreaHeader = fmt.Sprintf(
			"\n\nInput (type or use voice): -- Generating for [%s]\n",
			renderedTime,
		)
	} else {
		textAreaHeader = "\n\nInput (type or use voice):\n"
	}

	textArea := m.textArea.View()

	return pager + textAreaHeader + textArea
}

func formatDuration(d time.Duration) string {
	// Limit to 60 seconds
	if d > 60*time.Second {
		d = 60 * time.Second
	}

	// Round to nearest intval
	d = d.Round(stopwatchIntval)

	// Format to ensure one decimal place
	seconds := d.Seconds()
	return fmt.Sprintf("%.1fs", seconds)
}
