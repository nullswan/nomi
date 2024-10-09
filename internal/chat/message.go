package chat

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	Id        uuid.UUID `json:"id"`
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	// TODO(nullswan): Add files
}

func NewMessage(role Role, content string) Message {
	return Message{
		Id:        uuid.New(),
		Role:      role,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}
}
