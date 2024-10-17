package code

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type MockExecutor struct {
	output string
	err    string
	code   int
}

func (m *MockExecutor) Execute(code string) ExecutionResult {
	return ExecutionResult{
		Stdout:   m.output,
		Stderr:   m.err,
		ExitCode: m.code,
	}
}

func TestExecuteCodeBlock(t *testing.T) {
	t.Parallel()

	// Register mock executors
	registerExecutor(
		"mock",
		&MockExecutor{output: "Mock output", err: "", code: 0},
	)
	registerExecutor(
		"error",
		&MockExecutor{output: "", err: "Mock error", code: 1},
	)

	tests := []struct {
		name     string
		block    CodeBlock
		expected ExecutionResult
	}{
		{
			name:  "Successful execution",
			block: CodeBlock{Language: "mock", Code: "test code"},
			expected: ExecutionResult{
				Stdout:   "Mock output",
				Stderr:   "",
				ExitCode: 0,
			},
		},
		{
			name:  "Execution with error",
			block: CodeBlock{Language: "error", Code: "test code"},
			expected: ExecutionResult{
				Stdout:   "",
				Stderr:   "Mock error",
				ExitCode: 1,
			},
		},
		{
			name:  "Unsupported language",
			block: CodeBlock{Language: "unsupported", Code: "test code"},
			expected: ExecutionResult{
				Stdout:   "",
				Stderr:   "Unsupported language: unsupported",
				ExitCode: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ExecuteCodeBlock(tt.block)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf(
					"ExecuteCodeBlock() = %v, want %v",
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestOsascriptExecution(t *testing.T) {
	t.Parallel()

	registerExecutor(
		"osascript",
		&MockExecutor{output: "Osascript output", err: "", code: 0},
	)

	block := CodeBlock{Language: "osascript", Code: "test code"}
	result := ExecuteCodeBlock(block)

	if runtime.GOOS == "darwin" {
		if result.Stdout != "Osascript output" {
			t.Errorf("Expected 'Osascript output', got %s", result.Stdout)
		}
	} else {
		if !strings.Contains(result.Stderr, "Osascript is only supported on macOS") {
			t.Errorf("Expected error about macOS support, got %s", result.Stderr)
		}
	}
}
