package code

import (
	"bufio"
	"strings"
)

func ParseCodeBlocks(input string) []CodeBlock {
	var blocks []CodeBlock
	scanner := bufio.NewScanner(strings.NewReader(input))
	var currentBlock CodeBlock
	inCodeBlock := false
	var selectedLang string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				if currentBlock.Language == selectedLang {
					blocks = append(blocks, currentBlock)
				}
				currentBlock = CodeBlock{}
				inCodeBlock = false
			} else {
				currentBlock.Language = strings.TrimSpace(strings.TrimPrefix(line, "```"))
				if selectedLang == "" {
					selectedLang = currentBlock.Language
				}
				inCodeBlock = true
			}
		} else if inCodeBlock {
			if currentBlock.Code == "" {
				currentBlock.Code = line
			} else {
				currentBlock.Code += "\n" + line
			}
		}
	}

	if inCodeBlock {
		if currentBlock.Language == selectedLang {
			blocks = append(blocks, currentBlock)
		}
	}

	return blocks
}
