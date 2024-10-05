package chat

import prompts "github.com/nullswan/golem/internal/prompt"

type Conversation interface {
	GetId() string
	GetMessages() []Message
	AddMessage(message Message)
	WithPrompt(prompt prompts.Prompt)
}
