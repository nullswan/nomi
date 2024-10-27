package copywriter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nullswan/nomi/internal/chat"
	"github.com/nullswan/nomi/internal/tools"
)

const outlineAgentPrompt = `Create multiple detailed outline plans for copywritten content based on provided goals, ideas, and a headline.

Take the provided goals into strong consideration as they are the foundation for the outline plans. Use the supplied ideas and headline to formulate detailed outlines that meet the user's requirements and input.

# Steps

1. **Analyze Input:**
   - Review the provided goals to understand the main objectives.
   - Note the ideas and headline as they will inform the context and direction of the content.

2. **Develop Outline Plans:**
   - Create a detailed table of contents.
   - Ensure outline plan aligns closely with the provided goals, ideas, and headline.
   - Include both first-level titles and, if necessary, subtitles or paragraph summaries to enhance detail and structure.

3. **Iterate on User Needs:**
   - Confirm that the outline plan is fitting to the user's stated needs and match the input specifications.

# Output Format

Return the response as a JSON object containing a plan. A plan should be detailed, including a table of contents with first-level titles and any needed subtitles or paragraph resumes.

# Examples

### Example Input:
- Goals: ["Increase brand awareness", "Engage young adults"]
- Ideas: ["Sustainable fashion", "Social media influence"]
- Headline: "The Future of Fashion: Sustainability Meets Style"

### Example Output:
{
		"table_of_contents": [
				{
						"section_title": "Introduction to Sustainable Fashion",
						"subsections": [
								{
										"title": "Importance and Impact",
										"summary": "Discuss the significance and environmental benefits of sustainable fashion."
								}
						]
				},
				{
						"section_title": "Combining Style with Sustainability",
						"subsections": [
								{
										"title": "Innovative Approaches",
										"summary": "Explore how brands are integrating style with ecological practices."
								}
						]
				},
				{
						"section_title": "Role of Social Media",
						"subsections": [
								{
										"title": "Influencer Impact",
										"summary": "Detail the role of social media influencers in promoting sustainable fashion."
								}
						]
				}
		]
}

# Notes

- Ensure that each outline plan is comprehensive and logically structured.
- Adapt the complexity and depth of the table of contents based on the nature of the topic and input goals.
- Confirm alignment with users' specific needs and suggestions outlined in the input.`

// This agent is responsible for defining the content plan
// Capabilities include: Define content plan, highlight key points
// Uses: goalsAgent, ideasAgent, headlineAgent
type outlineAgent struct {
	logger            tools.Logger
	textToJSONBackend tools.TextToJSONBackend
	selector          tools.Selector
	inputHandler      tools.InputHandler

	storage outlinePlan

	goalsAgent    *goalsAgent
	ideasAgent    *ideasAgent
	headlineAgent *headlineAgent
}

func NewOutlineAgent(
	logger tools.Logger,
	textToJSONBackend tools.TextToJSONBackend,
	selector tools.Selector,
	inputHandler tools.InputHandler,
	goalsAgent *goalsAgent,
	ideasAgent *ideasAgent,
	headlineAgent *headlineAgent,
) *outlineAgent {
	return &outlineAgent{
		logger:            logger,
		textToJSONBackend: textToJSONBackend,
		selector:          selector,
		inputHandler:      inputHandler,
		goalsAgent:        goalsAgent,
		ideasAgent:        ideasAgent,
		headlineAgent:     headlineAgent,
	}
}

func (o *outlineAgent) GetStorage() string {
	ret := ""
	for _, toc := range o.storage.TableOfContents {
		ret += toc.SectionTitle + "\n"
		for _, sub := range toc.Subsections {
			ret += "\t" + sub.Title + ": " + sub.Summary + "\n"
		}
	}

	return ret
}

func (c *outlineAgent) OnStart(
	ctx context.Context,
	conversation chat.Conversation,
) error {
	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleSystem),
			outlineAgentPrompt,
		),
	)

	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleUser),
			"Goals:\n"+c.goalsAgent.GetStorage(),
		),
	)

	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleUser),
			"Headline:\n"+c.headlineAgent.GetStorage(),
		),
	)

	conversation.AddMessage(
		chat.NewMessage(
			chat.Role(chat.RoleUser),
			"Ideas:\n"+c.ideasAgent.GetStorage(),
		),
	)

	for done := false; !done; {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		default:
			resp, err := c.textToJSONBackend.Do(ctx, conversation)
			if err != nil {
				return fmt.Errorf("error generating completion: %w", err)
			}

			conversation.AddMessage(
				chat.NewMessage(
					chat.Role(chat.RoleAssistant),
					resp,
				),
			)

			var outlineResp outlinePlan
			if err := json.Unmarshal([]byte(resp), &outlineResp); err != nil {
				return fmt.Errorf("error unmarshalling response: %w", err)
			}

			if len(outlineResp.TableOfContents) == 0 {
				return fmt.Errorf("outline plan is empty")
			}

			fmt.Println("Outline Plan:")
			for _, toc := range outlineResp.TableOfContents {
				fmt.Println("\t" + toc.SectionTitle)
				for _, sub := range toc.Subsections {
					fmt.Println("\t\t" + sub.Title + ": " + sub.Summary)
				}
			}

			if !c.selector.SelectBool(
				"Is the outline plan satisfactory?",
				true,
			) {
				newInstructions, err := c.inputHandler.Read(ctx, ">>> ")
				if err != nil {
					return fmt.Errorf("error reading new instructions: %w", err)
				}

				conversation.AddMessage(
					chat.NewMessage(
						chat.Role(chat.RoleUser),
						newInstructions,
					),
				)
				continue
			}

			done = true
			c.storage = outlineResp
		}
	}

	fmt.Println("Outline Plan agent completed")
	return nil
}

type outlinePlan struct {
	TableOfContents []tableOfContents `json:"table_of_contents"`
}

type tableOfContents struct {
	SectionTitle string       `json:"section_title"`
	Subsections  []subsection `json:"subsections"`
}

type subsection struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}
