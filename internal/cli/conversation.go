package cli

import (
	"fmt"

	"github.com/nullswan/nomi/internal/chat"
	prompts "github.com/nullswan/nomi/internal/prompt"
)

// InitConversation initializes the conversation.
// If conversationID is nil, a new conversation is created.
func InitConversation(
	repo chat.Repository,
	conversationID *string,
	defaultPrompt prompts.Prompt,
) (chat.Conversation, error) {
	var err error
	var conversation chat.Conversation

	if conversationID == nil || *conversationID == "" {
		conversation = chat.NewStackedConversation(repo)
		conversation.WithPrompt(defaultPrompt)
		return conversation, nil
	}

	conversation, err = repo.LoadConversation(*conversationID)
	if err != nil {
		return nil, fmt.Errorf("error loading conversation: %w", err)
	}

	return conversation, nil
}
