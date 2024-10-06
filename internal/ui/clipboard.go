package ui

import "github.com/atotto/clipboard"

func getClipboard() (string, error) {
	return clipboard.ReadAll()
}

func setClipboard(text string) error {
	return clipboard.WriteAll(text)
}
