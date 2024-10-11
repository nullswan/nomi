package code

import (
	"reflect"
	"testing"
)

func TestParseCodeBlocks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []CodeBlock
	}{
		{
			name:  "Multiple languages",
			input: "```python\nprint('Hello')\n```\n```bash\necho 'World'\n```",
			expected: []CodeBlock{
				{Language: "python", Code: "print('Hello')"},
				{Language: "bash", Code: "echo 'World'"},
			},
		},
		{
			name:     "No code blocks",
			input:    "This is just plain text",
			expected: []CodeBlock{},
		},
		{
			name:  "Empty code block",
			input: "```python\n```",
			expected: []CodeBlock{
				{Language: "python", Code: ""},
			},
		},
		{
			name:  "Unclosed code block",
			input: "```python\nprint('Unclosed')",
			expected: []CodeBlock{
				{Language: "python", Code: "print('Unclosed')"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ParseCodeBlocks(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseCodeBlocks() = %v, want %v", result, tt.expected)
			}
		})
	}
}
