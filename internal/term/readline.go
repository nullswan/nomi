package term

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ollama/ollama/readline"
)

var (
	ErrInputInterrupted = errors.New("input interrupted")
	ErrInputKilled      = errors.New("input killed")
	ErrReadlineInit     = errors.New("error initializing readline")
)

func InitReadline() (*readline.Instance, error) {
	rl, err := readline.New(readline.Prompt{
		Prompt:      ">>> ",
		AltPrompt:   "...  ",
		Placeholder: "Send a message (/help for help)",
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReadlineInit, err)
	}
	return rl, nil
}

type MultilineState int

const (
	MultilineNone MultilineState = iota
	MultilinePrompt
	MultilineSystem
	MultilineTemplate
)

func ReadInput(
	inputCh chan<- string,
	inputErrCh chan<- error,
	readyCh chan struct{},
) {
	defer close(readyCh)

	// Wait the readyCh once before starting the loop
	_, ok := <-readyCh
	if !ok {
		inputErrCh <- ErrInputInterrupted
		return
	}

	pipedInput, err := getPipedInput()
	if err != nil {
		inputErrCh <- fmt.Errorf("error reading piped input: %w", err)
		return
	}

	if pipedInput != "" {
		fmt.Println(">>>", pipedInput)
		inputCh <- pipedInput
	}

	rl, err := InitReadline()
	if err != nil {
		inputErrCh <- fmt.Errorf("%w: %v", ErrReadlineInit, err)
		return
	}

	fmt.Print(readline.StartBracketedPaste)
	defer fmt.Printf(readline.EndBracketedPaste)

	var sb strings.Builder
	var multiline MultilineState

	for {
		line, err := rl.Readline()
		switch {
		case errors.Is(err, io.EOF):
			fmt.Println()
			inputErrCh <- ErrInputKilled
			return
		case errors.Is(err, readline.ErrInterrupt):
			if line == "" {
				fmt.Println("\nUse CTRL+D or /exit to exit.")
			}

			rl.Prompt.UseAlt = false
			sb.Reset()

			continue
		case err != nil:
			inputErrCh <- fmt.Errorf("error reading input: %w", err)
			return
		}

		switch {
		case multiline != MultilineNone:
			// check if there's a multiline terminating string
			before, ok := strings.CutSuffix(line, `"""`)
			sb.WriteString(before)
			if !ok {
				fmt.Fprintln(&sb)
				continue
			}

			multiline = MultilineNone
			rl.Prompt.UseAlt = false
		case strings.HasPrefix(line, `"""`):
			line := strings.TrimPrefix(line, `"""`)
			line, ok := strings.CutSuffix(line, `"""`)
			sb.WriteString(line)
			if !ok {
				// no multiline terminating string; need more input
				fmt.Fprintln(&sb)
				multiline = MultilinePrompt
				rl.Prompt.UseAlt = true
			}
			continue
		case rl.Pasting:
			fmt.Fprintln(&sb, line)
			continue
		default:
			sb.WriteString(line)
		}

		if sb.Len() > 0 && multiline == MultilineNone {
			inputCh <- sb.String()
			sb.Reset()
		}

		// Wait for the readyCh before reading the next line
		_, ok := <-readyCh
		if !ok {
			inputErrCh <- ErrInputInterrupted
			return
		}
	}
}
