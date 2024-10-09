package term

import (
	"bufio"
	"fmt"
	"os"
)

func NewInputArea() string {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	emptyLines := 0

	for {
		fmt.Fprint(os.Stderr, ">>> ")
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()

		if text == "" {
			emptyLines++
		} else {
			emptyLines = 0
		}

		lines = append(lines, text)

		if emptyLines >= 2 {
			break
		}
	}

	// Erase the printed lines
	for i := 0; i < len(lines); i++ {
		fmt.Fprint(os.Stderr, "\033[1A\033[2K") // Move up and clear line
	}

	// Remove the last empty line
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Return the input as a single string
	result := ""
	for _, line := range lines {
		result += line + "\n"
	}

	return result
}
