package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

type KnowledgeBase interface {
	QueryAll() (string, error)
	Close() error
}

type FileKnowledgeBase struct {
	repositoryPath string
}

func NewFileKnowledgeBase(repositoryPath string) *FileKnowledgeBase {
	return &FileKnowledgeBase{
		repositoryPath: repositoryPath,
	}
}

func (f *FileKnowledgeBase) QueryAll() (string, error) {
	files, err := os.ReadDir(f.repositoryPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var content string
	for _, file := range files {
		if !file.IsDir() {
			data, err := os.ReadFile(
				filepath.Join(f.repositoryPath, file.Name()),
			)
			if err != nil {
				return "", fmt.Errorf(
					"failed to read file %s: %w",
					file.Name(),
					err,
				)
			}
			content += string(data)
		}
	}

	return content, nil
}

func (f *FileKnowledgeBase) Close() error {
	return nil
}
