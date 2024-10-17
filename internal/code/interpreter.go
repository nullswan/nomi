package code

func InterpretCodeBlocks(input string) []ExecutionResult {
	blocks := ParseCodeBlocks(input)
	results := make([]ExecutionResult, len(blocks))

	for i, block := range blocks {
		results[i] = ExecuteCodeBlock(block)
	}

	return results
}
