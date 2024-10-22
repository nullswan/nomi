package term

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func getPipedInput() (string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("error checking stdin stat: %v", err)
	}

	if stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0 {
		return "", nil
	}

	reader := bufio.NewReader(os.Stdin)
	var b strings.Builder

	for {
		r, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}

		_, err = b.WriteRune(r)
		if err != nil {
			return "", fmt.Errorf("error writing rune to buffer: %v", err)
		}
	}

	return strings.TrimSpace(b.String()), nil
}
