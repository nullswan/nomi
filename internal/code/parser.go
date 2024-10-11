package code

import (
	"bufio"
	"strings"
)

func ParseCodeBlocks(input string) []CodeBlock {
	blocks := []CodeBlock{}
	scanner := bufio.NewScanner(strings.NewReader(input))
	var currentBlock CodeBlock
	inCodeBlock := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				blocks = append(blocks, currentBlock)
				currentBlock = CodeBlock{}
				inCodeBlock = false
			} else {
				currentBlock.Language = strings.TrimSpace(strings.TrimPrefix(line, "```"))
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
		blocks = append(blocks, currentBlock)
	}

	return blocks
}
