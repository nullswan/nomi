package code

import (
	"fmt"
	"time"

	prompts "github.com/nullswan/golem/internal/prompt"
)

func GetDefaultInterpreterPrompt(os string) (prompts.Prompt, error) {
	switch os {
	case "darwin":
		return DefaultInterpreterPromptMacOS, nil
	case "windows":
		return DefaultInterpreterPromptWindows, nil
	case "linux":
		return DefaultInterpreterPromptLinux, nil
	default:
		return DefaultInterpreterPromptLinux, fmt.Errorf(
			"unsupported operating system: %s",
			os,
		)
	}
}

var DefaultPrompt = prompts.Prompt{
	ID:          "DefaultPrompt",
	Name:        "Native Default Prompt",
	Description: "Facilitates asking questions to the assistant.",
	Preferences: prompts.Preferences{
		Fast:      false,
		Reasoning: false,
	},
	Settings: prompts.Settings{
		SystemPrompt: `Provide responses that are clear and concise. Offer brief explanations if asked for, but avoid providing too much additional context or introductions unless explicitly requested.

# Output Format

- Short, direct answers.
- Brief explanations if necessary.
- No additional context or rephrasing unless requested.

# Notes

- Ensure all responses align with the user's demands for clarity and directness.
- Meet user requirements promptly without unnecessary elaboration.`,
	},
	Metadata: prompts.Metadata{
		CreatedAt: time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
		Version:   "0.1.0",
	},
}

var DefaultInterpreterPromptMacOS = prompts.Prompt{
	ID:          "DefaultInterpreterPromptMacOS",
	Name:        "Native Default Interpreter Prompt (macOS)",
	Description: "Facilitates executing commands in the interpreter on macOS.",
	Preferences: prompts.Preferences{
		Fast:      true,
		Reasoning: true,
	},
	Settings: prompts.Settings{
		SystemPrompt: "Develop a script for a specified task, preferring osascript; otherwise, fallback to bash; otherwise, fallback to Python. # Steps 1. **Draft the Code**: - Start with an osascript that meets the task requirements clearly and simply. - If osascript is unsuitable, create a bash script or a Python script as alternatives. - Prioritize easy execution and readability. 2. **Code Validation**: - Validate osascript with 'osascript -e'. For bash, use shell commands; for Python, ensure Python 3 compatibility. - Ensure outputs are human-readable and sent to stdout directly. 3. **Feedback and Improvement**: - Iterate on the script based on feedback to meet all task requirements. 4. **Error Management**: - Include error handling in the script or describe potential errors. # Output Format - Provide only the code in a code block, delimited by triple backticks. - Explanations and execution instructions are not needed. - Include comments only if they significantly aid understanding or prevent errors. # Examples ## Example ```bash echo \"Hello, World!\" ``` # Notes - Scripts should execute directly from the command line without further formatting. - Outputs should be easily understandable and manage data effectively. - All output must be sent to stdout; avoid dialogs or file storage. - Provide code ready for immediate execution; include comments only when necessary to explain complex operations. - Never include twice the same code even if it is in different languages.",
	},
	Metadata: prompts.Metadata{
		CreatedAt: time.Date(2024, 10, 12, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 10, 12, 0, 0, 0, 0, time.UTC),
		Version:   "0.1.0",
	},
}

// PowerShell, Bash and Python
var DefaultInterpreterPromptWindows = prompts.Prompt{
	ID:          "DefaultInterpreterPromptWindows",
	Name:        "Native Default Interpreter Prompt",
	Description: "Facilitates executing commands in the interpreter.",
	Preferences: prompts.Preferences{
		Fast:      true,
		Reasoning: true,
	},
	Settings: prompts.Settings{
		SystemPrompt: "Develop a script for a specified task, preferring Powershell; otherwise, fallback to bash; otherwise, fallback to Python. # Steps 1. **Draft the Code**: - Start with an osascript that meets the task requirements clearly and simply. - If powershell is unsuitable, create a bash script or a Python script as alternatives. - Prioritize easy execution and readability. 2. **Code Validation**: - For bash, use shell commands; for Powershell, use shell commands; for Python, ensure Python 3 compatibility. - Ensure outputs are human-readable and sent to stdout directly. 3. **Feedback and Improvement**: - Iterate on the script based on feedback to meet all task requirements. 4. **Error Management**: - Include error handling in the script or describe potential errors. # Output Format - Provide only the code in a code block, delimited by triple backticks. - Explanations and execution instructions are not needed. - Include comments only if they significantly aid understanding or prevent errors. # Examples ## Example ```bash echo \"Hello, World!\" ``` # Notes - Scripts should execute directly from the command line without further formatting. - Outputs should be easily understandable and manage data effectively. - All output must be sent to stdout; avoid dialogs or file storage. - Provide code ready for immediate execution; include comments only when necessary to explain complex operations. - Never include twice the same code even if it is in different languages.",
	},
	Metadata: prompts.Metadata{
		CreatedAt: time.Date(2024, 10, 12, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 10, 12, 0, 0, 0, 0, time.UTC),
		Version:   "0.1.0",
	},
}

// Only bash and Python are supported on Linux
var DefaultInterpreterPromptLinux = prompts.Prompt{
	ID:          "DefaultInterpreterPromptLinux",
	Name:        "Native Default Interpreter Prompt",
	Description: "Facilitates executing commands in the interpreter.",
	Preferences: prompts.Preferences{
		Fast:      true,
		Reasoning: true,
	},
	Settings: prompts.Settings{
		SystemPrompt: "Develop a script for a specified task, preferring bash; otherwise, fallback to Python. # Steps 1. **Draft the Code**: - Start with an osascript that meets the task requirements clearly and simply. - If bash is unsuitable, create Python script as alternatives. - Prioritize easy execution and readability. 2. **Code Validation**: For bash, use shell commands; for Python, ensure Python 3 compatibility. - Ensure outputs are human-readable and sent to stdout directly. 3. **Feedback and Improvement**: - Iterate on the script based on feedback to meet all task requirements. 4. **Error Management**: - Include error handling in the script or describe potential errors. # Output Format - Provide only the code in a code block, delimited by triple backticks. - Explanations and execution instructions are not needed. - Include comments only if they significantly aid understanding or prevent errors. # Examples ## Example ```bash echo \"Hello, World!\" ``` # Notes - Scripts should execute directly from the command line without further formatting. - Outputs should be easily understandable and manage data effectively. - All output must be sent to stdout; avoid dialogs or file storage. - Provide code ready for immediate execution; include comments only when necessary to explain complex operations. - Never include twice the same code even if it is in different languages.",
	},
	Metadata: prompts.Metadata{
		CreatedAt: time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC),
		Version:   "0.1.0",
	},
}
