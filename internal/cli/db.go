package cli

import (
	"fmt"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/code"
)

// InitChatDatabase initializes the chat repository.
func InitChatDatabase(sqlitePath string) (chat.Repository, error) {
	repo, err := chat.NewSQLiteRepository(sqlitePath)
	if err != nil {
		return nil, fmt.Errorf("error creating repository: %w", err)
	}
	return repo, nil
}

// InitCodeDatabase initializes the code repository.
func InitCodeDatabase(sqlitePath string) (code.Repository, error) {
	repo, err := code.NewSQLiteRepository(sqlitePath)
	if err != nil {
		return nil, fmt.Errorf("error creating repository: %w", err)
	}
	return repo, nil
}
