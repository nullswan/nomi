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

var DefaultInterpreterInferencePrompt = prompts.Prompt{
	ID:          "DefaultInterpreterInferencePrompt",
	Name:        "Default Interpreter Inference Prompt",
	Description: "Facilitates executing commands in the interpreter.",
	Preferences: prompts.Preferences{
		Fast:      true,
		Reasoning: false,
	},
	Settings: prompts.Settings{
		SystemPrompt: "Classify the purpose of a given piece of code by providing a concise yet accurate description.\nEnsure your description is clear and precise, as it will help in selecting the appropriate script to execute.\n# Steps\n1. **Analyze the Code**: Read and understand the code snippet provided.\n2. **Determine Functionality**: Identify the primary functionality and objective of the code.\n3. **Craft the Description**: Write a brief and precise description of what the code is doing.\n# Output Format\nThe output should be a short sentence clearly describing the main function or action of the code snippet.\n# Examples\n**Input:** \n```python\ndef add_numbers(a, b):\n    return a + b\n```\n**Output:** \n\"The code adds two numbers.\"\n**Input:** \n```javascript\ndocument.querySelector('button').addEventListener('click', function() {\n    alert('Hello, world!');\n});\n```\n**Output:** \n\"The code adds a click event listener to a button to display an alert.\"\n# Notes\n- Focus on the main action the code is performing.\n- Be concise but ensure the description is descriptive enough to make an informed decision.",
	},
	Metadata: prompts.Metadata{
		CreatedAt: time.Date(2024, 10, 13, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 10, 13, 0, 0, 0, 0, time.UTC),
		Version:   "0.1.0",
	},
}

var DefaultInterpreterCachePrompt = prompts.Prompt{
	ID:          "DefaultInterpreterCachePrompt",
	Name:        "Default Interpreter Cache Prompt",
	Description: "Facilitates executing commands in the interpreter.",
	Preferences: prompts.Preferences{
		Fast:      true,
		Reasoning: false,
	},
	// Might want to use the ID there.
	Settings: prompts.Settings{
		SystemPrompt: "Identify whether a user's question relates to any code block in cache by comparing the question to a list of code block IDs and their descriptions. If the question addresses the exact same problem as any code block in the cache, return the code block's ID. If it does not match any code block, indicate that you don't know. # Steps 1. **Input Parsing**: Receive a structured list of pairs, each containing a `codeblock_id` and its `description`. 2. **Question Comparison**: Analyze the user's question and compare it to the descriptions of the cached code blocks. 3. **Matching Identification**: Check if the question is directly related to the problem described in any of the code blocks. 4. **Output Decision**: - If a match is found, output the corresponding `codeblock_id`. - If no match is found, output \"I don't know.\" # Output Format Provide the output as a simple text response: - \"I don't know\" if no match is found. - \"CODEBLOCK_ID\" if a match is found, where `CODEBLOCK_ID` is the identifier of the matching code block. # Examples **Example 1** - **Input**: - Cached code blocks: - (\"codeblock_id\": \"001\", \"description\": \"Calculates the factorial of a number\") - (\"codeblock_id\": \"002\", \"description\": \"Sorts a list of integers in ascending order\") - User question: \"How do I sort numbers in a list?\" - **Output**: \"002\" **Example 2** - **Input**: - Cached code blocks: - (\"codeblock_id\": \"003\", \"description\": \"Parses JSON data\") - (\"codeblock_id\": \"004\", \"description\": \"Connects to a database\") - User question: \"How to compress a file in Python?\" - **Output**: \"I don't know\"",
	},
	Metadata: prompts.Metadata{
		CreatedAt: time.Date(2024, 10, 13, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 10, 13, 0, 0, 0, 0, time.UTC),
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
		SystemPrompt: "Develop a script for a specified task, preferring osascript; otherwise, fallback to bash; otherwise, fallback to Python. # Steps 1. **Draft the Code**: - Start with an osascript that meets the task requirements clearly and simply. - If osascript is unsuitable, create a bash script or a Python script as alternatives. - Prioritize easy execution and readability. 2. **Code Validation**: - Validate osascript with 'osascript -e'. For bash, use shell commands; for Python, ensure Python 3 compatibility. - Ensure outputs are human-readable and sent to stdout directly. 3. **Feedback and Improvement**: - Iterate on the script based on feedback to meet all task requirements. 4. **Error Management**: - Include error handling in the script or describe potential errors. # Output Format - Provide only the code in a code block, delimited by triple backticks. - Explanations and execution instructions are not needed. - Include comments only if they significantly aid understanding or prevent errors. # Examples ## Example ```osascript\ntell application \"Calendar\"\nend tell\n``` ## Example ```bash echo \"Hello, World!\" ``` # Notes - Scripts should execute directly from the command line without further formatting. - Outputs should be easily understandable and manage data effectively. - All output must be sent to stdout; avoid dialogs or file storage. - Provide code ready for immediate execution; include comments only when necessary to explain complex operations. - Never include twice the same code even if it is in different languages.",
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
