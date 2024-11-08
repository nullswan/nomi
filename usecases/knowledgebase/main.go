package knowledgebase

import (
	"context"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/tools"
)

func OnStart(
	ctx context.Context,
	selector tools.Selector,
	logger tools.Logger,
	tttProvider tools.TextToTextBackend,
	conversation chat.Conversation,
) error {
	return nil
}
