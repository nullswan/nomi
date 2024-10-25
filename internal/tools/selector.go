package tools

import "github.com/nullswan/nomi/internal/term"

type Selector interface {
	SelectBool(title string, defaultValue bool) bool
}

type selector struct{}

func NewSelector() Selector {
	return &selector{}
}

func (s *selector) SelectBool(title string, defaultValue bool) bool {
	return term.PromptForBool(title, defaultValue)
}
