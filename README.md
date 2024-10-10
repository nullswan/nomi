# Golem

A simple, boring, and productive LLM runtime.

> **Note:** This project is still in development and not yet ready for production use.
>
> However, feel free to try it out and provide feedback.
> 
> You can email me at [pro@nullswan.io](mailto:pro@nullswan.io) or open an issue on GitHub.

# Table of Contents

- [Features](#features)
  - [Conversation Management](#conversation-management)
  - [Prompt Management](#prompt-management)
  - [AI Provider Integration](#ai-provider-integration)
  - [Terminal Input Handling](#terminal-input-handling)
  - [Markdown Output](#markdown-output)
  - [Fallback on Auto Ollama Installation](#fallback-on-auto-ollama-installation)
- [Example Usage](#example-usage)
  - [Basic Conversation](#basic-conversation)
  - [Cancelation](#cancelation)
  - [Prompt Management](#prompt-management-1)
  - [File Upload](#file-upload)

## Features

### Conversation Management
- Create, load, and list conversations.
- Supports adding messages to conversations.
- Handles file messages.

### Prompt Management
- Add, edit, and list prompts.
- Fetch prompts from specified URLs.
- Validate prompts before saving.

### AI Provider Integration
- Integration with multiple AI providers (OpenAI, Ollama).
- Generate text completions based on user input.
- Support for streaming responses from AI models.

### Terminal Input Handling
- Supports interactive terminal input.
- Can read piped input from standard input.

### Markdown Output
- Render AI responses in a formatted markdown style.
- Incorporate emojis and styled output for better readability.

### Fallback on Auto Ollama Installation
- Automatically downloads and installs necessary binaries for Ollama when not already present.

## Example Usage

### Basic Conversation

```shell
golem
>>> Hello, World!
>>>
>>>
AI:
Hello, World!
```

### Cancelation

```shell
golem
>>> Give me a list of 100 random words.
>>>
>>>
AI:
Here is a list of 100 random words:
1. Word1
2. Word2^CRequest canceled by user.
```

### Prompt Management

```shell
golem prompt list
>>> ...

golem prompt add https://raw.githubusercontent.com/nullswan/golem/refs/heads/main/prompts/native-code.yml
>>> Prompt added successfully.

golem prompt edit native-code
...
```
### File Upload

```shell
golem
>>> ./users/root/file.txt
>>>
>>>
Added file: ./users/root/file.txt

>>> What is the content of the file?
>>>
>>>

AI:
The content of the file is:
Hello, World!
```