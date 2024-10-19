package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/nullswan/nomi/internal/chat"
)

func HandleCommands(text string, conversation chat.Conversation) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return text
	}

	ret := ""
	for _, line := range lines {
		if !strings.HasPrefix(line, "/") {
			ret += line + "\n"
			continue
		}

		switch {
		case strings.HasPrefix(line, "/help"):
			printHelp()
		case strings.HasPrefix(line, "/reset"):
			conversation = conversation.Reset()
			fmt.Println("Conversation reset.")
		case strings.HasPrefix(line, "/add"):
			args := strings.Fields(line)
			if len(args) < 2 {
				fmt.Println("Usage: /add <file or directory>")
				continue
			}

			if !isLocalResource(args[1]) {
				fmt.Println("Invalid file or directory: " + args[1])
				continue
			}

			processLocalResource(conversation, args[1])
		case strings.HasPrefix(line, "/exit"):
			fmt.Println("Exiting...")
			os.Exit(0)
		default:
			fmt.Println("Unknown command:", line)
			printHelp()
		}
	}

	return ret
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  /help        Show this help message")
	fmt.Println("  /reset       Reset the conversation")
	fmt.Println(
		"  /add <file>  Add a file or directory to the conversation",
	)
	fmt.Println("  /exit        Exit the application")
}
