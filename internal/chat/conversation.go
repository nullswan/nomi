package chat

import (
	"time"

	"github.com/google/uuid"
	prompts "github.com/nullswan/nomi/internal/prompt"
)

type Conversation interface {
	GetID() string
	GetCreatedAt() time.Time
	GetMessages() []Message

	RemoveMessage(id uuid.UUID)
	AddMessage(message Message)

	WithPrompt(prompt prompts.Prompt)

	Reset() Conversation
}
