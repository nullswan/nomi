package code

import (
	"testing"
)

func TestFormatExecutionResultForLLM(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		results  []ExecutionResult
		expected string
	}{
		{
			name: "Single result with all fields",
			results: []ExecutionResult{
				{
					Stdout:   "Hello, World!",
					Stderr:   "Warning: This is a test",
					ExitCode: 1,
				},
			},
			expected: `--- Execution Result 1 ---

Error:
Warning: This is a test

Output:
Hello, World!

Exit Code: 1`,
		},
		{
			name: "Multiple results with varying fields",
			results: []ExecutionResult{
				{
					Stdout: "First output",
				},
				{
					Stderr:   "Second error",
					ExitCode: 2,
				},
				{
					Stdout:   "Third output",
					Stderr:   "Third error",
					ExitCode: 0,
				},
			},
			expected: `--- Execution Result 1 ---

Output:
First output

----------------------------------------

--- Execution Result 2 ---

Error:
Second error

Exit Code: 2

----------------------------------------

--- Execution Result 3 ---

Error:
Third error

Output:
Third output`,
		},
		{
			name:     "Empty results",
			results:  []ExecutionResult{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FormatExecutionResultForLLM(tt.results)
			if result != tt.expected {
				t.Errorf(
					"FormatExecutionResultForLLM() = %v, want %v",
					result,
					tt.expected,
				)
			}
		})
	}
}
