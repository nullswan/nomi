package ollamaprovider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/ollama/ollama/api"
	"golang.org/x/sync/errgroup"
)

func stopOllamaServer(
	cmd *exec.Cmd,
) error {
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		return fmt.Errorf(
			"error sending interrupt signal to ollama server: %w",
			err,
		)
	}

	// Wait for 3 seconds for the process to exit gracefully then kill it by force.
	stop := make(chan any, 1)
	wg := errgroup.Group{}

	wg.Go(func() error {
		err := cmd.Wait()
		stop <- struct{}{}
		return fmt.Errorf(
			"error waiting for ollama server to stop: %w",
			err,
		)
	})

	wg.Go(func() error {
		select {
		case <-stop:
		case <-time.After(3 * time.Second):
			// Ignore errors if there is a race
			// and the process has already closed.
			_ = cmd.Process.Kill()
		}
		return nil
	})

	err := wg.Wait()
	if err != nil {
		return fmt.Errorf("error stopping ollama server: %w", err)
	}

	return nil
}

// https://github.com/redpanda-data/connect/blob/f1786c54e132691d37ccb1bd8041c5658886eb6f/internal/impl/ollama/base_processor.go
func waitForOllamaServer(client *api.Client) error {
	timeout := time.After(5 * time.Second)
	tick := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-timeout:
			return errors.New("timed out waiting for server to start")
		case <-tick.C:
			if err := client.Heartbeat(context.Background()); err == nil {
				return nil // server has started
			}
		}
	}
}
