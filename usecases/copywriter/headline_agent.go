package copywriter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/tools"
)

const headlineAgentPrompt = `Generate at least 10 creative and varied headlines for copywriting content based on the user's initial goals and ideas.

- Consider the user's specific goals and ideas carefully to tailor the headlines.
- Ensure variety in format and style across the headlines.
- Aim for creativity and originality in each suggestion.

# Steps

1. Review and comprehend the user's initial goals and ideas.
2. Brainstorm diverse headlines that align with the user's objectives.
3. Vary the formats and styles used to maintain creativity across headlines.
4. Compile at least 10 unique headlines.

# Output Format

Return the content in JSON format with a key "headlines," which contains an array of headline strings:
{
  "headlines": [
    "Headline 1",
    "Headline 2",
    ...
    "Headline 10"
  ]
}

# Examples

Example input:
- User Goal: Increase engagement for a blog about healthy eating.
- User Idea: Highlight quick and easy recipes.

Example output:
{
  "headlines": [
    "Quick and Easy: 5-Minute Recipes for Health Enthusiasts",
    "Transform Your Meals: Easy Healthy Recipes to Try Today",
    "Say Goodbye to Long Cooking: Quick Healthy Dishes",
    "5 Simple Recipes That Boost Your Health Right Now",
    "Healthy Eating Made Easy: Quick Fix Meals",
    "Easy yet Delicious: Healthy Dishes for Busy Lives",
    "Fast and Wholesome: Your Guide to Healthy Eating",
    "Harness the Power of Quick Healthy Meals",
    "Minimal Cook Time, Maximum Health: Recipe Guide",
    "Discover the Ease of Healthy Eating with These Recipes"
  ]
}

# Notes

- Use diverse language structures and rhetorical devices to enhance creativity.
- Ensure all headlines align with the input theme and purpose.
- Adapt headlines if user provides feedback for revision.`

// This agent is responsible for defining the headline of the content
// Capabilities include: Define headline
// Uses: goalsAgent, ideasAgent
type headlineAgent struct {
	logger            tools.Logger
	textToJSONBackend tools.TextToJSONBackend
	selector          tools.Selector
	inputArea         tools.InputArea

	storage string

	goalsAgent *goalsAgent
	ideasAgent *ideasAgent
}

func NewHeadlineAgent(
	logger tools.Logger,
	textToJSONBackend tools.TextToJSONBackend,
	selector tools.Selector,
	inputArea tools.InputArea,
	goalsAgent *goalsAgent,
	ideasAgent *ideasAgent,
) *headlineAgent {
	return &headlineAgent{
		logger:            logger,
		textToJSONBackend: textToJSONBackend,
		selector:          selector,
		inputArea:         inputArea,
		goalsAgent:        goalsAgent,
		ideasAgent:        ideasAgent,
	}
}

const NoneOfTheAboveChoice = "None of the above"

func (h *headlineAgent) OnStart(
	ctx context.Context,
	conversation chat.Conversation,
) error {
	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleSystem),
			headlineAgentPrompt,
		),
	)
	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleUser),
			"Goals:\n"+h.goalsAgent.GetStorage(),
		),
	)
	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleUser),
			"Ideas:\n"+h.ideasAgent.GetStorage(),
		),
	)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			resp, err := h.textToJSONBackend.Do(ctx, conversation)
			if err != nil {
				return fmt.Errorf("error generating completion: %w", err)
			}

			conversation.AddMessage(
				chat.NewMessage(
					chat.Role(chat.RoleAssistant),
					resp,
				),
			)

			var headlineResp headlineResponse
			if err := json.Unmarshal([]byte(resp), &headlineResp); err != nil {
				return fmt.Errorf("error unmarshalling response: %w", err)
			}

			if headlineResp.Headlines == nil {
				return fmt.Errorf("headline result is empty")
			}

			ret := h.selector.SelectString(
				"Choose your headline",
				append(headlineResp.Headlines, NoneOfTheAboveChoice),
			)

			if ret == NoneOfTheAboveChoice {
				response, err := h.inputArea.Read(">>> ")
				if err != nil {
					return fmt.Errorf("error getting input: %w", err)
				}

				conversation.AddMessage(
					chat.NewMessage(
						chat.Role(chat.RoleUser),
						response,
					),
				)
				continue
			}

			h.logger.Info("Selected headline: " + ret)
			h.storage = ret
			return nil
		}
	}
}

func (h *headlineAgent) GetStorage() string {
	return h.storage
}

type headlineResponse struct {
	Headlines []string `json:"headlines"`
}
