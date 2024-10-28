package term

// From: https://github.com/ollama/ollama/blob/main/readline/readline.go

import (
	"bufio"
	"io"
	"os"
)

type Terminal struct {
	outchan chan rune
	done    chan struct{}
	rawmode bool
	termios any
	reader  *bufio.Reader
}

func NewTerminal() (*Terminal, error) {
	t := &Terminal{
		outchan: make(chan rune),
		done:    make(chan struct{}),
		rawmode: false,
		termios: nil,
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
