package textinput

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Be able to read from stdin
// Be able to open a VIM editor
// Be able to cancel the ongoing prompt
func Run(ctx context.Context, outCh chan string) {
	var wg sync.WaitGroup

	// Read from stdin
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
		}
	}()

	// Wait for stdin to close
	wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Print("\n> ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			input = strings.TrimSpace(input)
		}
	}
}
