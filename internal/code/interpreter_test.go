package code

import (
	"reflect"
	"testing"
)

func TestInterpretCodeBlocks(t *testing.T) {
	t.Parallel()

	// Register mock executors
	registerExecutor(
		"python",
		&MockExecutor{output: "Python output", err: "", code: 0},
	)
	registerExecutor(
		"bash",
		&MockExecutor{output: "Bash output", err: "", code: 0},
	)

	input := "```python\nprint('Hello')\n```\n```bash\necho 'World'\n```"
	expected := []ExecutionResult{
		{Stdout: "Python output", Stderr: "", ExitCode: 0},
		{Stdout: "Bash output", Stderr: "", ExitCode: 0},
	}

	results := InterpretCodeBlocks(input)

	if !reflect.DeepEqual(results, expected) {
		t.Errorf("InterpretCodeBlocks() = %v, want %v", results, expected)
	}
}
