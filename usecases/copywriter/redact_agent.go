package copywriter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/tools"
)

const redactAgentPrompt = `You are tasked with generating content for a project based on given inputs. You will receive goals, ideas, headlines, and outlines, and your objective is to create multiple markdown-formatted documents optimized for specific platforms such as LinkedIn or a blog.

# Steps

1. **Understand the Inputs:**
   - **Goals:** Identify the purpose and target audience of the content. Determine platform requirements (LinkedIn, Blog, etc.).
   - **Ideas**: Analyze the provided ideas to form the core message and themes.
   - **Headline**: Craft a compelling starting point that aligns with the goals.
   - **Outline**: Use the outline to structure the content logically.

2. **Content Creation:**
   - Draft the markdown documents adhering to the outline.
   - Optimize for the specified platform (e.g., SEO optimization for a blog, concise and engaging language for LinkedIn).
   - Ensure content is easy to read, using simple language and avoiding overly rare words.

3. **Feedback and Revision:**
   - Review for coherence, structure, and human readability.
   - Adjust any sections that do not align with the goals, platform optimization needs, or readability guidelines.

# Output Format

- Produce a JSON object containing an array of documents. Each element in the array represents a document formatted in markdown. Ensure JSON syntax correctness and markdown compliance.

# Examples

### Example Format:
Input:
{
  "goals": ["Increase brand awareness", "Engage LinkedIn audience"],
  "ideas": ["Highlight company values", "Introduce new product feature"],
  "headline": "Exciting Innovations at [CompanyName]",
  "outline": ["Introduction", "Main Features", "Conclusion"]
}
Output:
[
  {
    "platform": "LinkedIn",
    "content": "# Exciting Innovations at [CompanyName]\n\n## Introduction\n[Write an engaging hook related to company values and highlight their importance.]\n\n## Main Features\n[Discuss the new product feature, highlighting its benefits and uniqueness.]\n\n## Conclusion\n[Wrap up with a call to action, encouraging engagement or contact.]"
  }
]
(Note: Replace placeholders with specific content. Length and detail should be consistent with typical LinkedIn post structure.)

# Notes

- Always tailor content based on the primary platform specified in the goals.
- Maintain coherence and ensure the content reflects a logical flow and human readability.
- For multi-platform goals, harmonize content to suit each platform’s unique style and requirements.`

// This agent is responsible for redacting the content
// Capabilities include: Redact, Edit, Proofread
// Uses: contentPlanAgent, toneStyleAgent
type redactAgent struct {
	logger            tools.Logger
	inputHandler      tools.InputHandler
	textToJSONBackend tools.TextToJSONBackend
	selector          tools.Selector

	exportAgent   *exportAgent
	outlineAgent  *outlineAgent
	headlineAgent *headlineAgent
	ideasAgent    *ideasAgent
	goalsAgent    *goalsAgent
}

func NewRedactAgent(
	goalsAgent *goalsAgent,
	ideasAgent *ideasAgent,
	headlineAgent *headlineAgent,
	outlineAgent *outlineAgent,
	exportAgent *exportAgent,
	logger tools.Logger,
	inputHandler tools.InputHandler,
	textToJSONBackend tools.TextToJSONBackend,
	selector tools.Selector,
) *redactAgent {
	return &redactAgent{
		logger:            logger,
		inputHandler:      inputHandler,
		exportAgent:       exportAgent,
		outlineAgent:      outlineAgent,
		headlineAgent:     headlineAgent,
		ideasAgent:        ideasAgent,
		goalsAgent:        goalsAgent,
		textToJSONBackend: textToJSONBackend,
		selector:          selector,
	}

}

func (r *redactAgent) OnStart(
	ctx context.Context,
	conversation chat.Conversation,
) error {
	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleSystem),
			redactAgentPrompt,
		),
	)

	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleUser),
			"Goals:\n"+r.goalsAgent.GetStorage(),
		),
	)

	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleUser),
			"Ideas:\n"+r.ideasAgent.GetStorage(),
		),
	)

	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleUser),
			"Headline:\n"+r.headlineAgent.GetStorage(),
		),
	)

	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleUser),
			"Outline:\n"+r.outlineAgent.GetStorage(),
		),
	)

	// TODO(nullswan): Generate part by part according to the outline
	// TODO(nullswan): Generate for each format independently

	for {
		select {
		case <-ctx.Done():
			return nil

		default:
			resp, err := r.textToJSONBackend.Do(ctx, conversation)
			if err != nil {
				return fmt.Errorf("error generating completion: %w", err)
			}

			conversation.AddMessage(
				chat.NewMessage(
					chat.Role(chat.RoleAssistant),
					resp,
				),
			)

			var redactResp redactResponse
			if err := json.Unmarshal([]byte(resp), &redactResp); err != nil {
				return fmt.Errorf("error unmarshalling response: %w", err)
			}

			for _, doc := range redactResp.Documents {
				r.logger.Info(
					"Platform: " + doc.Platform,
				)

				r.logger.Info(
					"Content: " + doc.Content,
				)
			}

			if r.selector.SelectBool(
				"Do you want to export the content to a file?",
				true,
			) {
				for _, doc := range redactResp.Documents {
					r.logger.Info(
						"Exporting " + doc.Platform + " content to file...",
					)
					err := r.exportAgent.ExportToFile(doc.Content)
					if err != nil {
						return fmt.Errorf(
							"error exporting content to file: %w",
							err,
						)
					}
				}

				return nil
			}

			resp, err = r.inputHandler.Read(ctx, ">>> ")
			if err != nil {
				return fmt.Errorf("error getting input: %w", err)
			}

			conversation.AddMessage(
				chat.NewMessage(
					chat.Role(chat.RoleUser),
					resp,
				),
			)
		}
	}
}

type document struct {
	Platform string `json:"platform"`
	Content  string `json:"content"`
}

type redactResponse struct {
	Documents []document `json:"documents"`
}
