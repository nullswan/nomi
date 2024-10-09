package chat

import (
	"fmt"
	"time"

	prompts "github.com/nullswan/golem/internal/prompt"
)

type stackedConversation struct {
	repo Repository

	id        string
	messages  []Message
	createdAt time.Time
}

// #region Getters
func (c *stackedConversation) GetId() string {
	return c.id
}

func (c *stackedConversation) GetCreatedAt() time.Time {
	return c.createdAt
}

func (c *stackedConversation) GetMessages() []Message {
	return c.messages
}

// #endregion

func (c *stackedConversation) AddMessage(message Message) {
	c.messages = append(c.messages, message)
	c.repo.SaveConversation(c)
}

func (c *stackedConversation) WithPrompt(prompt prompts.Prompt) {
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

func NewStackedConversation(
	repo Repository,
) Conversation {
	id := fmt.Sprintf( // TODO(nullswan): Use configurable ID format
		"sc_%d",
		time.Now().Unix(),
	)

	return &stackedConversation{
		repo:      repo,
		id:        id,
		messages:  make([]Message, 0),
		createdAt: time.Now(),
	}
}
