package copywriter

import (
	"context"

	"github.com/nullswan/nomi/internal/chat"
	baseprovider "github.com/nullswan/nomi/internal/providers/base"
	"github.com/nullswan/nomi/internal/tools"
)

// TODO(nullswan): Add Memory, Storage, Tools
// TODO(nullswan): Generate multiple content examples

// Pull preferences
// Ask Goal
// Ask and retrieve inputs
// Define headline
// Define/Gather ideas
// Define Content Plan
// Define tone & style
// Redact
// Export

// This agent is responsible for defining the goals of the content
// Capabilities include: Define goals, Gather goals
type goalsAgent struct {
	textBackend  *baseprovider.TextToTextProvider
	systemPrompt string
}

func NewGoalsAgent() *goalsAgent {
	return &goalsAgent{
		systemPrompt: "",
	}
}

// This agent is responsible for defining the content ideas
// Capabilities include: Define ideas, Gather ideas
type ideasAgent struct {
	textBackend  *baseprovider.TextToTextProvider
	systemPrompt string
}

// This agent is responsible for defining the headline of the content
// Capabilities include: Define headline
// Uses: goalsAgent, ideasAgent
type headlineAgent struct {
	textBackend  *baseprovider.TextToTextProvider
	systemPrompt string
}

// This agent is responsible for defining the content plan
// Capabilities include: Define content plan, highlight key points
// Uses: goalsAgent, ideasAgent, headlineAgent
type contentPlanAgent struct {
	textBackend  *baseprovider.TextToTextProvider
	systemPrompt string
}

// This agent is responsible for defining the tone and style of the content
// Capabilities include: Define tone, Define style
// Uses: goalsAgent, headlineAgent
type toneStyleAgent struct {
	textBackend  *baseprovider.TextToTextProvider
	systemPrompt string
}

// This agent is responsible for redacting the content
// Capabilities include: Redact, Edit, Proofread
// Uses: contentPlanAgent, toneStyleAgent
type redactAgent struct {
	textBackend  *baseprovider.TextToTextProvider
	systemPrompt string
}

// This agent is responsible for exporting the final content
// Capabilities include: Export to file, export to different formats
type exportAgent struct {
	textBackend  *baseprovider.TextToTextProvider
	toolBackend  *baseprovider.TextToTextProvider
	systemPrompt string
}

func OnStart(
	ctx context.Context,
	selectors tools.Selector,
	logger tools.Logger,
	inputArea tools.InputArea,
	textToTextBackend tools.TextToTextBackend,
	conversation chat.Conversation,
) {
}
