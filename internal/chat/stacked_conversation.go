package chat

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	prompts "github.com/nullswan/nomi/internal/prompt"
)

type stackedConversation struct {
	repo Repository

	id        string
	messages  []Message
	createdAt time.Time
}

// #region Getters
func (c *stackedConversation) GetID() string {
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
	err := c.repo.SaveConversation(c)
	if err != nil {
		fmt.Println(err)
	}
}

func (c *stackedConversation) RemoveMessage(id uuid.UUID) {
	for i, message := range c.messages {
		if message.ID == id {
			c.messages = append(c.messages[:i], c.messages[i+1:]...)
			break
		}
	}
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

// TODO(nullswan): Conversation should remain immutable
func (c *stackedConversation) Reset() (Conversation, error) {
	err := c.repo.SaveConversation(c)
	if err != nil {
		return nil, fmt.Errorf("error saving conversation: %w", err)
	}

	conversation := NewStackedConversation(c.repo)

	// Copy system messages
	for _, message := range c.messages {
		if message.Role != RoleSystem {
			break
		}
		conversation.AddMessage(
			message,
		)
	}

	c.createdAt = conversation.GetCreatedAt()
	c.id = conversation.GetID()
	c.messages = conversation.GetMessages()

	return c, nil
}

func (c *stackedConversation) Clean() (Conversation, error) {
	err := c.repo.SaveConversation(c)
	if err != nil {
		return nil, fmt.Errorf("error saving conversation: %w", err)
	}

	conversation := NewStackedConversation(c.repo)

	c.createdAt = conversation.GetCreatedAt()
	c.id = conversation.GetID()
	c.messages = conversation.GetMessages()

	return c, nil
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
