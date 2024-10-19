package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nullswan/nomi/internal/chat"
)

func isLocalResource(text string) bool {
	var path string
	if strings.HasPrefix(text, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		path = filepath.Join(home, text[1:])
	} else {
		path = text
	}

	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "/") {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

func processLocalResource(conversation chat.Conversation, text string) {
	if isDirectory(text) {
		addAllFiles(conversation, text)
	} else {
		addSingleFile(conversation, text)
	}
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func addAllFiles(conversation chat.Conversation, directory string) {
	files, err := os.ReadDir(directory)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}
	for _, file := range files {
		path := filepath.Join(directory, file.Name())
		if file.IsDir() {
			addAllFiles(conversation, path)
		} else {
			addFileToConversation(conversation, path, file.Name())
		}
	}
}

func addSingleFile(conversation chat.Conversation, filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	fileName := filepath.Base(filePath)
	conversation.AddMessage(
		chat.NewFileMessage(
			chat.RoleUser,
			formatFileMessage(fileName, string(content)),
		),
	)
	fmt.Printf("Added file: %s\n", filePath)
}

func addFileToConversation(
	conversation chat.Conversation,
	filePath, fileName string,
) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	conversation.AddMessage(
		chat.NewFileMessage(
			chat.RoleUser,
			formatFileMessage(fileName, string(content)),
		),
	)
	fmt.Printf("Added file: %s\n", filePath)
}

func formatFileMessage(fileName, content string) string {
	return fileName + "-----\n" + content + "-----\n"
}
