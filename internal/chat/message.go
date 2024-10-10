package chat

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID `json:"id"`
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	IsFile    bool      `json:"is_file"`
}

func NewMessage(role Role, content string) Message {
	return Message{
		ID:        uuid.New(),
		Role:      role,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}
}

func NewFileMessage(role Role, content string) Message {
	return Message{
		ID:        uuid.New(),
		Role:      role,
		Content:   content,
		CreatedAt: time.Now().UTC(),
		IsFile:    true,
	}
}
