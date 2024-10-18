# Nomi - A Multi-Modal AI Runtime

Born from the principles that **data ownership is challenging** and **AI simplifies communication**, the Nomi project aims to complement human intelligence by orchestrating data to enhance **decision-making** and handle **automated micro-tasks**.

Our mission is to make AI more accessible and easy to use for everyone, ensuring that you have control over your data whenever you need it.

**Nomi**, is a versatile and lightweight AI runtime that integrates with AI services like OpenAI, OpenRouter, and Ollama, enabling you to leverage their capabilities. While using these providers involves sending your data externally, Nomi also maintains a local archive of your data, ensuring you retain control and ownership.

Supporting multiple data types, including text and audio, with image support coming soon, Nomi offers flexibility while safeguarding your information.

> **Note:** This project is still being developed and isn't ready for full use yet, but we are working hard to make it stable and reliable.
>
> We welcome any feedback, suggestions, or contributions. Thank you for trying Nomi!

# Table of Contents

- [Features](#features)
  - [Manage Conversations](#manage-conversations)
  - [Manage Prompts](#manage-prompts)
  - [AI Provider Integration](#ai-provider-integration)
  - [Terminal Features](#terminal-features)
  - [Real-time Voice](#real-time-voice)
  - [Interpreter Mode](#interpreter-mode)
  - [Additional Features](#additional-features)
- [Roadmap](#roadmap)
- [Installation](#installation)


## Nomi

![Nomi Logo](./nomi-logo.png)


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

## Installation

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
