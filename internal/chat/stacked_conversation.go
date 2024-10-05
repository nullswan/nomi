package chat

import (
	"fmt"
	"time"
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
