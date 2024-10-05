package prompts

import "time"

var DefaultPrompt = Prompt{
	ID:          "DefaultPrompt",
	Name:        "Native Default Prompt",
	Description: "Facilitates asking questions to the assistant.",
	Settings: Settings{
		SystemPrompt: `Provide responses that are clear and concise. Offer brief explanations if asked for, but avoid providing too much additional context or introductions unless explicitly requested.

# Output Format

- Short, direct answers.
- Brief explanations if necessary.
- No additional context or rephrasing unless requested.

# Notes

- Ensure all responses align with the user's demands for clarity and directness.
- Meet user requirements promptly without unnecessary elaboration.`,
	},
	Metadata: Metadata{
		CreatedAt: time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
		Version:   "0.1.0",
	},
}
