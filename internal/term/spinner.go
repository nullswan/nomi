package term

import (
	"fmt"
	"time"
)

const spinnerRefreshRate = 50 * time.Millisecond

type Spinner struct {
	stop            chan bool
	frames          []string
	initialDuration time.Duration
	initialMessage  string
}

func NewSpinner(initialDuration time.Duration, initialMessage string) *Spinner {
	frames := []string{
		"⠋",
		"⠙",
		"⠹",
		"⠸",
		"⠼",
		"⠴",
		"⠦",
		"⠧",
		"⠇",
		"⠏",
	}
	return &Spinner{
		stop:            make(chan bool),
		frames:          frames,
		initialDuration: initialDuration,
		initialMessage:  initialMessage,
	}
}

func (s *Spinner) Start() {
	// Clear the line
	fmt.Print("\033[2K\r")

	go func() {
		i := 0
		start := time.Now()
		for {
			select {
			case <-s.stop:
				if elapsed := time.Since(start); elapsed < s.initialDuration {
					fmt.Printf("\r%s", s.initialMessage)
				} else {
					fmt.Print("\r")
				}
				return
			default:
				elapsed := time.Since(start)
				var color string
				reset := "\033[0m"
				if elapsed < s.initialDuration {
					color = "\033[31m" // Red
				} else {
					color = "\033[32m" // Green
				}
				fmt.Printf("\r%s%s%s", color, s.frames[i%len(s.frames)], reset)
				i++
				time.Sleep(spinnerRefreshRate)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.stop <- true
}
