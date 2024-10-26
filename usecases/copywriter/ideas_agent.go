package copywriter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/tools"
)

const ideasAgentPrompt = `Your task is to process and clarify a collection of raw ideas, which are provided as a series of entries. These entries may not follow any specific format and may consist of brief thoughts. Your role is to interpret these ideas and ask necessary clarifying questions in JSON format, facilitating a better understanding of the user's aims. The refined ideas will ultimately be used for crafting copywriting content.

# Steps

1. **Receive Input:**
   - You will receive multiple idea entries, each separated by triple backticks in the input.
   - You will also receive the goals of the copywriting content to provide context.

2. **Interpret Ideas:**
   - Review each idea entry.
   - Determine the core message or intent of the idea.

3. **Ask Clarifying Questions:**
   - Formulate questions to gain clarity on ambiguous or incomplete ideas.
   - Ensure each question aims to understand the user’s intentions more clearly.

4. **JSON Format for Questions:**
   - Use a JSON structure with the following keys:
     - "question": Your formulated question.
     - "done": Boolean value indicating whether more questions are needed. Initially set this to false.
     - "result": Fill this in only when all questions have been answered and no further clarification is required.

5. **Review and Finalize:**
   - Once clarity is achieved on an idea, set the "done" key to true.
   - Summarize the refined idea in the "result" key.

# Output Format

Respond with JSON objects for each idea entry. Each object should contain the keys "question", "done", and "result". Initially fill in "question" and "done", and once clarification is achieved, complete the "result".

# Examples

**Input:**
Basketball game promotion, focus on team spirit.

**Output:**
{
  "question": "Can you specify the key elements you associate with team spirit?",
  "done": false,
  "result": ""
}

**Input:**
Eco-friendly packaging for beauty product line, consumer awareness.

**Output:**
{
  "question": "What specific consumer actions do you want to encourage with this eco-friendly packaging?",
  "done": false,
  "result": ""
}

# Notes

- If an idea is already clear and does not require further questions, set "done" to true and summarize in "result".
- Focus on ensuring the resulting questions lead directly to the attainment of the copywriting goals provided.`

// This agent is responsible for defining the content ideas
// Capabilities include: Define ideas, Gather ideas
type ideasAgent struct {
	logger            tools.Logger
	inputArea         tools.InputArea
	textToJSONBackend tools.TextToJSONBackend

	storage string

	goalsAgent *goalsAgent
}

func NewIdeasAgent(
	logger tools.Logger,
	textToJSONBackend tools.TextToJSONBackend,
	goalsAgent *goalsAgent,
	inputArea tools.InputArea,
) *ideasAgent {
	return &ideasAgent{
		textToJSONBackend: textToJSONBackend,
		logger:            logger,
		goalsAgent:        goalsAgent,
		inputArea:         inputArea,
	}
}

func (i *ideasAgent) GetStorage() string {
	return i.storage
}

func (i *ideasAgent) OnStart(
	ctx context.Context,
	conversation chat.Conversation,
) error {
	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleSystem),
			ideasAgentPrompt,
		),
	)

	var ideas []string

	// TODO(nullswan): The empty line behavior is not very manageable
	i.logger.Info("Gathering ideas, push empty line to finish")
	for done := false; !done; {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		default:
			idea, err := i.inputArea.Read(">>> ")
			if err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}

			if idea == "/done" {
				done = true
				break
			}

			ideas = append(ideas, idea)
		}
	}

	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleUser),
			"Goals:\n"+i.goalsAgent.GetStorage(),
		),
	)

	for n, idea := range ideas {
		conversation.AddMessage(
			chat.NewMessage(
				chat.Role(chat.RoleUser),
				fmt.Sprintf("Idea %d\n```\n%s\n```", n+1, idea),
			),
		)
	}

	for done := false; !done; {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		default:
			resp, err := i.textToJSONBackend.Do(ctx, conversation)
			if err != nil {
				return fmt.Errorf("error generating completion: %w", err)
			}

			conversation.AddMessage(
				chat.NewMessage(
					chat.Role(chat.RoleAssistant),
					resp,
				),
			)

			var ideaResp ideaResponse
			if err := json.Unmarshal([]byte(resp), &ideaResp); err != nil {
				return fmt.Errorf("error unmarshalling response: %w", err)
			}

			if ideaResp.Done {
				if ideaResp.Result == "" {
					return fmt.Errorf("idea result is empty")
				}

				i.storage = ideaResp.Result

				done = true
				break
			}

			fmt.Println("Next question: ", ideaResp.Question)
			response, err := i.inputArea.Read(">>> ")
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

	i.logger.Info("Ideas agent completed")
	i.logger.Info("Ideas summary: " + i.storage)

	return nil
}

type ideaResponse struct {
	Question string `json:"question"`
	Done     bool   `json:"done"`
	Result   string `json:"result"`
}
