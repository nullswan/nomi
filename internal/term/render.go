package term

import (
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

type Renderer = glamour.TermRenderer

func InitRenderer() (*glamour.TermRenderer, error) {
	var style ansi.StyleConfig
	switch {
	case !term.IsTerminal(int(os.Stdout.Fd())):
		style = styles.NoTTYStyleConfig
	case termenv.HasDarkBackground():
		style = styles.DarkStyleConfig
	default:
		style = styles.LightStyleConfig
	}
	style.Document.Margin = new(uint)
	style.CodeBlock.Margin = new(uint)

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(0),
		glamour.WithEmoji(),
		glamour.WithPreservedNewLines(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating renderer: %w", err)
	}

	return renderer, nil
}
