package chat

type Conversation interface {
	GetId() string
	GetMessages() []Message
	AddMessage(message Message)
}
