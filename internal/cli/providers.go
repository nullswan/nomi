package cli

import (
	"fmt"

	"github.com/nullswan/nomi/internal/logger"
	"github.com/nullswan/nomi/internal/providers"
	baseprovider "github.com/nullswan/nomi/internal/providers/base"
)

// InitProviders initializes the text-to-text provider.
func InitProviders(
	logger *logger.Logger,
	targetModel string,
	reasoning bool,
) (baseprovider.TextToTextProvider, error) {
	provider := providers.CheckProvider()

	var textToTextBackend baseprovider.TextToTextProvider
	if reasoning {
		var err error
		textToTextBackend, err = providers.LoadTextToTextReasoningProvider(
			provider,
			targetModel,
		)
		if err != nil {
			logger.
				With("error", err).
				Error("Error loading text-to-text reasoning provider")
		}
	}
	if textToTextBackend == nil {
		var err error
		textToTextBackend, err = providers.LoadTextToTextProvider(
			provider,
			targetModel,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"error loading text-to-text provider: %w",
				err,
			)
		}
	}

	return textToTextBackend, nil
}
