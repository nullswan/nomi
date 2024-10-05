package chat

import "time"

type Message struct {
	Role      Role   `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	// TODO(nullswan): Add files
}

func NewMessage(role Role, content string) Message {
	return Message{
		Role:      role,
		Content:   content,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
}
