package term

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func NewInputArea() string {
	var ttyPath string
	if runtime.GOOS == "windows" {
		ttyPath = "CON"
	} else {
		ttyPath = "/dev/tty"
	}

	tty, err := os.Open(ttyPath)
	if err != nil {
		fmt.Println("Error opening terminal:", err)
		return ""
	}
	defer tty.Close()

	scanner := bufio.NewScanner(tty)
	var lines []string
	emptyLines := 0

	for {
		fmt.Fprint(os.Stdout, ">>> ")
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
		fmt.Fprint(os.Stdout, "\033[1A\033[2K") // Move up and clear line
	}

	// Remove the last empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n")
}
