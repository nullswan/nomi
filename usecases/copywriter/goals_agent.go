package copywriter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/tools"
)

const goalsAgentPrompt = `Determine and clarify precise goals, corner cases, usage, format, and objectives for copywriting projects by asking one question at a time. Respond in JSON format.

Ensure that you respond in JSON with the following structure:
- 'question': The question you want to ask regarding the copywriting project.
- 'next': A boolean indicating whether you have more questions. Set this to 'false' if there are no more questions.
- 'result': Provide a detailed and concise summary of the goals you defined once all questions are answered and next is set to 'false'.

If you have any questions, follow this structure when formulating them:
- Begin with "Question [current number]/[total number]:" and include the question.
- Continue asking until all necessary details are gathered.

# Output Format

The output should be formatted as JSON in the following structure:

{
  "question": "[Your question here]",
  "next": true // or false if there are no more questions
}

# Examples

**Example JSON with a question:**

{
  "question": "Question 1/3: What is the main objective of the copywriting project?",
  "next": true
}

**Example JSON when no more questions are needed:**

{
  "question": "",
	"result": "A detailed and concise summary of the goals you defined by asking questions.",
  "next": false
}

**Example JSON summary upon instruction:**

"A detailed and concise summary of the goals you defined by asking questions."

# Notes

- An user can request multiple formats, tones, or styles for the copywriting project.
- Make sure each question is detailed enough to extract specific information.
- Clarify corner cases and considerations relevant to the copywriting project goals.
- The summary should encapsulate all relevant goals and objectives gathered through questions.
- You can ask as many questions as necessary. If you have more questions or receive ambiguous answers, you can change the initial question plan and ask additional questions, even if you planned a specific number of questions previously.
- When asking question, be as precise as possible to ensure the goals are well-defined, you can provide a few examples to clarify the question.`

// This agent is responsible for defining the goals of the content
// Capabilities include: Define goals, Gather goals
type goalsAgent struct {
	textToJSONBackend tools.TextToJSONBackend
	inputArea         tools.InputArea
	logger            tools.Logger
	selector          tools.Selector

	storage string
}

func NewGoalsAgent(
	textToJSONBackend tools.TextToJSONBackend,
	inputArea tools.InputArea,
	logger tools.Logger,
	selector tools.Selector,
) *goalsAgent {
	return &goalsAgent{
		textToJSONBackend: textToJSONBackend,
		inputArea:         inputArea,
		selector:          selector,
		logger:            logger,
	}
}

func (g *goalsAgent) OnStart(
	ctx context.Context,
	conversation chat.Conversation,
) error {
	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleSystem),
			goalsAgentPrompt,
		),
	)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		default:
			resp, err := g.textToJSONBackend.Do(ctx, conversation)
			if err != nil {
				return fmt.Errorf("error generating completion: %w", err)
			}

			conversation.AddMessage(
				chat.NewMessage(
					chat.Role(chat.RoleAssistant),
					resp,
				),
			)

			var goalResp goalResponse
			if err := json.Unmarshal([]byte(resp), &goalResp); err != nil {
				return fmt.Errorf("error unmarshalling response: %w", err)
			}

			if !goalResp.Next {
				g.storage = goalResp.Result
				g.logger.Info(
					"Goals have been successfully defined!",
				)

				g.logger.Info(
					"Goals summary: " + goalResp.Result,
				)

				// TODO(nullswan): Ask user if they want to proceed with the goals

				return nil
			}

			fmt.Println("Next question: ", goalResp.Question)
			response, err := g.inputArea.Read(">>> ")
			if err != nil {
				return fmt.Errorf("error getting input: %w", err)
			}

			// Store the response
			conversation.AddMessage(
				chat.NewMessage(
					chat.Role(chat.RoleUser),
					response,
				),
			)
		}
	}
}

func (g goalsAgent) GetStorage() string {
	return g.storage
}

type goalResponse struct {
	Question string `json:"question"`
	Next     bool   `json:"next"`
	Result   string `json:"result"`
}
