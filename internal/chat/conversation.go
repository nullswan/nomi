package chat

import (
	"time"

	prompts "github.com/nullswan/nomi/internal/prompt"
)

type Conversation interface {
	GetID() string
	GetCreatedAt() time.Time
	GetMessages() []Message

	AddMessage(message Message)

	WithPrompt(prompt prompts.Prompt)

	Reset() Conversation
}
