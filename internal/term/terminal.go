package term

// From: https://github.com/ollama/ollama/blob/main/readline/readline.go

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/ollama/ollama/readline"
)

type Terminal struct {
	outchan chan rune
	done    chan struct{}
	rawmode bool
	termios any
	reader  *bufio.Reader
}

func NewTerminal() (*Terminal, error) {
	fd := os.Stdin.Fd()
	termios, err := readline.SetRawMode(fd)
	if err != nil {
		return nil, fmt.Errorf("failed to set raw mode: %w", err)
	}

	t := &Terminal{
		outchan: make(chan rune),
		done:    make(chan struct{}),
		rawmode: false,
		termios: termios,
		reader:  bufio.NewReader(os.Stdin),
	}

	go t.ioloop()

	return t, nil
}

func (t *Terminal) Read() (rune, error) {
	r, ok := <-t.outchan
	if !ok {
		return 0, io.EOF
	}

	return r, nil
}

func (t *Terminal) Close() error {
	close(t.outchan)
	close(t.done)
	if t.rawmode {
		readline.UnsetRawMode(os.Stdin.Fd(), t.termios)
	}
	return nil
}

func (t *Terminal) Closed() bool {
	select {
	case <-t.done:
		return true
	default:
		return false
	}
}

func (t *Terminal) ioloop() {
	defer func() {
		recover() // nolint:errcheck
	}()

	for {
		select {
		case <-t.done:
			return
		default:
			r, _, err := t.reader.ReadRune()
			if err != nil {
				return
			}

			select {
			case <-t.done:
				return
			case t.outchan <- r:
			}
		}
	}
}
