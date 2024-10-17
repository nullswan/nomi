package transcription

import (
	"context"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type TextReconciler struct {
	existingText string
	enableFixing bool
	apiKey       string
}

func NewTextReconciler(apiKey string) *TextReconciler {
	return &TextReconciler{
		apiKey: apiKey,
	}
}

func (tr *TextReconciler) SetEnableFixing(enabled bool) {
	tr.enableFixing = enabled
}

// Reconcile merges existing and new text, handling overlaps.
func (tr *TextReconciler) Reconcile(existing, newText string) string {
	trimmedNew := strings.TrimSpace(newText)
	if existing != "" && !strings.HasSuffix(existing, " ") {
		existing += " "
	}
	combined := existing + trimmedNew

	if tr.enableFixing {
		fixed, err := tr.FixText(combined)
		if err != nil {
			fmt.Printf("FixText error: %v\n", err)
			return combined
		}
		return fixed
	}

	return combined
}

// GetExistingText retrieves the current transcribed text.
func (tr *TextReconciler) GetExistingText() string {
	return tr.existingText
}

// UpdateText updates the existing transcribed text.
func (tr *TextReconciler) UpdateText(newText string) {
	tr.existingText = newText
}

// FixText implements text correction using OpenAI's ChatCompletion API.
func (tr *TextReconciler) FixText(text string) (string, error) {
	client := openai.NewClient(tr.apiKey)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prompt := fmt.Sprintf(
		"Please correct the following transcription for accuracy and readability:\n\n%s",
		text,
	)

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are an assistant that corrects and improves transcriptions for accuracy and readability.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("chat completion error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return text, fmt.Errorf("no response from OpenAI")
	}

	fixedText := strings.TrimSpace(resp.Choices[0].Message.Content)
	return fixedText, nil
}
