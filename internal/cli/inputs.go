package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
	"github.com/nullswan/nomi/internal/term"
)

func ReadInput(
	rl *readline.Instance,
	inputCh chan<- string,
	inputErrCh chan<- error,
) {
	pipedInput, err := term.GetPipedInput()
	if err != nil {
		inputErrCh <- fmt.Errorf("error reading piped input: %w", err)
		return
	}

	if pipedInput != "" {
		fmt.Println(">>>", pipedInput)
		inputCh <- pipedInput
	}

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				inputErrCh <- term.ErrInputInterrupted
				return
			}
			if err == io.EOF {
				// when killed, wait for alive..
				inputErrCh <- term.ErrInputKilled
				return
			}
			inputErrCh <- fmt.Errorf("error reading input: %w", err)
			return
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		inputCh <- line
	}
}
