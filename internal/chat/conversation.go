package chat

import (
	"time"

	"github.com/google/uuid"
	prompts "github.com/nullswan/nomi/internal/prompt"
)

type Conversation interface {
	// GetID returns the unique identifier of the conversation.
	GetID() string

	// GetCreatedAt returns the time when the conversation was created.
	GetCreatedAt() time.Time

	// GetMessages returns all messages in the conversation, ordered by creation date.
	GetMessages() []Message

	// RemoveMessage deletes a message from the conversation by its ID.
	RemoveMessage(id uuid.UUID)

	// AddMessage appends a new message to the conversation.
	AddMessage(message Message)

	// WithPrompt attaches a prompt to the conversation.
	WithPrompt(prompt prompts.Prompt)

	// Reset clears the conversation but retains system messages.
	Reset() (Conversation, error)

	// Clean removes all messages from the conversation, including system messages.
	Clean() (Conversation, error)
}
