package tools

import (
	"fmt"

	"github.com/nullswan/nomi/internal/term"
)

type InputArea interface {
	Read(defaultValue string) (string, error)
}

func NewInputArea() InputArea {
	return &inputArea{}
}

type inputArea struct{}

func (i *inputArea) Read(defaultValue string) (string, error) {
	input, err := term.ReadInputOnce()
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	return input, nil
}
