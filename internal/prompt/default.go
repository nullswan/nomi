package prompts

import "time"

var DefaultPrompts = []Prompt{
	{
		ID:          "commit_message",
		Name:        "Native Commit Message Prompt",
		Description: "Generates clear and concise commit messages.",
		Settings: Settings{
			SystemPrompt: "You are a helpful assistant for writing Git commit messages.",
			PrePrompt:    "Generate a commit message based on the following changes.",
		},
		Metadata: Metadata{
			CreatedAt: time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
			Version:   "0.1.0",
		},
	},
	{
		ID:          "ask",
		Name:        "Native Ask Prompt",
		Description: "Facilitates asking questions to the assistant.",
		Settings: Settings{
			SystemPrompt: "You are an informative assistant ready to answer questions.",
			PrePrompt:    "Awaiting the user's question.",
		},
		Metadata: Metadata{
			CreatedAt: time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
			Version:   "0.1.0",
		},
	},
	{
		ID:          "rephrase",
		Name:        "Native Rephrase Prompt",
		Description: "Rephrases user-provided text for clarity.",
		Settings: Settings{
			SystemPrompt: "You are a skilled assistant for rephrasing text.",
			PrePrompt:    "Rephrase the following text:",
		},
		Metadata: Metadata{
			CreatedAt: time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
			Version:   "0.1.0",
		},
	},
	{
		ID:          "summarize",
		Name:        "Native Summarize Prompt",
		Description: "Summarizes lengthy text into concise summaries.",
		Settings: Settings{
			SystemPrompt: "You are an expert assistant for summarizing content.",
			PrePrompt:    "Summarize the following text:",
		},
		Metadata: Metadata{
			CreatedAt: time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
			Version:   "0.1.0",
		},
	},
	{
		ID:          "code",
		Name:        "Native Code Prompt",
		Description: "Assists in writing and reviewing code snippets.",
		Settings: Settings{
			SystemPrompt: "You are a proficient assistant for coding tasks.",
			PrePrompt:    "Provide code for the following requirement:",
		},
		Metadata: Metadata{
			CreatedAt: time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
			Version:   "0.1.0",
		},
	},
}
