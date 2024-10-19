<p align="center">
  <img src="https://github.com/user-attachments/assets/5984c6d7-0d10-4325-af74-a69a87ff0eff" alt="Nomi, the interactive layer for AI">
</p>

> **Note:** This project is still being developed and isn't ready for full use yet, but we are working hard to make it stable and reliable.
>
> We welcome any feedback, suggestions, or contributions. Thank you for trying Nomi!

# Introduction

**Nomi** is a realtime multi-modal interface (i.e., voice and text interface) designed to **interact with your local data** and **automate your micro-tasks** thanks to LLM capabilities.

### Features

- **Multi-Modal Interface**: Accepts text or voice inputs (soon visual inputs)
- **Data Integration**: Interact with local files or public sources on internet
- **Task Automation**: Built-in code interpreter for generating and running code snippets
- **Conversation Management**: Organize, save, continue or reset conversations
- **Prompt Engineering**: Create, edit, and manage system prompts
- **Provider Flexibility**: Support local (e.g., Ollama) and cloud (e.g., OpenAI, OpenRouter) LLM providers

### Why Nomi?

In a world where data ownership is challenging and AI is revolutionizing communication, Nomi acts as a bridge between private data and LLM providers.
It supports integration with both local and external providers, including OpenAI, OpenRouter, and Ollama.

While using external providers, like OpenAI, involves transmitting data to third parties, Nomi also works with local providers such as Ollama.
Our aim is to democratize AI by making it more accessible and user-friendly, while ensuring control and ownership over data.

### Use cases

- **Personal AI Assistant**: Daily tasks automation and information retrieval
- **Code Development Aid**: Quick testing and coding assistance
- **Voice-Controlled AI**: Hands-free AI interaction
- **Local Data Analysis**: Privacy-focused data interaction
- **AI Model Comparison**: Easy switching between providers

# Get Started

### Linux & MacOS

```shell
# Check the script before running it.
curl -sSL https://raw.githubusercontent.com/nullswan/nomi/main/install.sh | bash
```

### Windows

```shell
# Check the script before running it.
curl -sSL https://raw.githubusercontent.com/nullswan/nomi/main/install.bat | cmd
```

### Compile from Source

```shell
git clone https://github.com/nullswan/nomi.git
cd nomi
make build
```

## Features

### Conversation Management

Create new conversations, load existing ones, list all conversations, add messages or files to a conversation, and reset conversations when needed.

### Prompts Management

Add new prompts, edit existing ones, list all available prompts, and fetch prompts from a URL for easy access and organization.

### Code interpreting capabilities

Run code on the fly with interpreter mode, making it easy to execute and test code directly within Nomi.

### AI Provider Integration

Connect seamlessly with AI providers like OpenAI, OpenRouter, and Ollama. Nomi can automatically install Ollama for you, simplifying the setup process.

### Real-time Voice

Enable real-time voice interactions, allowing for a more dynamic and interactive user experience.

### Terminal Experience

Enjoy markdown-formatted output, read inputs with piped commands, and cancel operations easily within the terminal.

### Additional Features

Explore more functionalities and upcoming features in the [Roadmap](#roadmap).

## Roadmap

These features are planned for future updates. They are not in any specific order and may be partially or not implemented yet.

- [ ] Image support
- [ ] Local Whisper
- [ ] Anthropic API
- [ ] Long transcript processing
- [ ] Transcript saving
- [ ] Daemon mode + HTTP Interface
- [ ] Editor mode
- [ ] Ask before making decisions
- [ ] Remember decisions
- [ ] Track metrics
- [ ] GUI
- [ ] Auto-update (Update command is already available)
- [ ] Interpreter safety
- [ ] Actions
- [ ] Action chains
- [ ] Real-time file management
- [ ] Memory embedding
