package term

import (
	"fmt"
	"io"
	"os"

	"github.com/muesli/cancelreader"
)

type Terminal struct {
	outchan chan rune
	done    chan struct{}
	rawmode bool
	termios any
	reader  cancelreader.CancelReader
}

func NewTerminal() (*Terminal, error) {
	reader, err := cancelreader.NewReader(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to create cancelable reader: %w", err)
	}
	t := &Terminal{
		outchan: make(chan rune),
		done:    make(chan struct{}),
		rawmode: false,
		termios: nil,
		reader:  reader,
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
	t.reader.Close()
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
	// This is where we recover from panics
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println("Recovered from panic:", rec)
		}
	}()

	for {
		select {
		case <-t.done:
			return
		default:
			buf := make([]byte, 1)
			n, err := t.reader.Read(buf)
			if err != nil {
				return
			}
			if n == 0 {
				continue
			}

			select {
			case <-t.done:
				return
			case t.outchan <- rune(buf[0]):
			}
		}
	}
}
