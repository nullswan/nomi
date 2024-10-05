package chat

import (
	"fmt"
	"time"

	prompts "github.com/nullswan/golem/internal/prompt"
)

type stackedConversation struct {
	messages []Message
	id       string
}

func (c *stackedConversation) GetId() string {
	return c.id
}

func (c *stackedConversation) GetMessages() []Message {
	return c.messages
}

func (c *stackedConversation) AddMessage(message Message) {
	c.messages = append(c.messages, message)
}

func (c stackedConversation) WithPrompt(prompt prompts.Prompt) {
	if prompt.Settings.SystemPrompt != "" {
		c.messages = append(c.messages, NewMessage(
			RoleSystem,
			prompt.Settings.SystemPrompt,
		))
	}

	if prompt.Settings.PrePrompt != nil && *prompt.Settings.PrePrompt != "" {
		c.messages = append(c.messages, NewMessage(
			RoleSystem,
			*prompt.Settings.PrePrompt,
		))
	}
}

func NewStackedConversation() Conversation {
	id := fmt.Sprintf( // TODO(nullswan): Use configurable ID format
		"sc_%d",
		time.Now().Unix(),
	)

	return &stackedConversation{
		id:       id,
		messages: make([]Message, 0),
	}
}
