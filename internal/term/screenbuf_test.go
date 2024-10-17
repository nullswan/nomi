package term

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

type mockWriter struct {
	buf bytes.Buffer
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	wN, wErr := m.buf.Write(p)
	if wErr != nil {
		return wN, fmt.Errorf("failed to write to buffer: %v", wErr)
	}

	return wN, nil
}

func TestNewScreenBuf(t *testing.T) {
	t.Parallel()

	w := &mockWriter{}
	sb := NewScreenBuf(w)

	if sb.writer != w {
		t.Errorf("Expected writer to be set correctly")
	}

	if sb.height > 22 || sb.height < 1 {
		t.Errorf("Expected height to be between 1 and 22, got %d", sb.height)
	}

	if cap(sb.lines) != sb.height {
		t.Errorf(
			"Expected lines capacity to be %d, got %d",
			sb.height,
			cap(sb.lines),
		)
	}
}

func TestWriteLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		lines    []string
		expected []string
	}{
		{
			name:     "Write fewer lines than height",
			lines:    []string{"Line 1", "Line 2", "Line 3"},
			expected: []string{"Line 1", "Line 2", "Line 3"},
		},
		{
			name:     "Write exactly height lines",
			lines:    []string{"1", "2", "3", "4", "5"},
			expected: []string{"1", "2", "3", "4", "5"},
		},
		{
			name:     "Write more than height lines",
			lines:    []string{"1", "2", "3", "4", "5", "6", "7"},
			expected: []string{"3", "4", "5", "6", "7"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := &mockWriter{}
			sb := NewScreenBuf(w)
			sb.height = 5 // Set a fixed height for testing

			for _, line := range tt.lines {
				sb.WriteLine(line)
			}

			if !equalSlices(sb.lines, tt.expected) {
				t.Errorf("Expected lines %v, got %v", tt.expected, sb.lines)
			}

			// Check if the output contains the expected lines
			output := w.buf.String()
			for _, line := range tt.expected {
				if !strings.Contains(output, line) {
					t.Errorf("Output doesn't contain expected line: %s", line)
				}
			}
		})
	}
}

func TestClear(t *testing.T) {
	t.Parallel()

	w := &mockWriter{}
	sb := NewScreenBuf(w)
	sb.height = 5 // Set a fixed height for testing

	lines := []string{"1", "2", "3", "4", "5"}
	for _, line := range lines {
		sb.WriteLine(line)
	}

	sb.Clear()

	if len(sb.lines) != 0 {
		t.Errorf("Expected lines to be empty after clear, got %v", sb.lines)
	}

	output := w.buf.String()
	expectedSequences := []string{
		"\033[5F", // Move up 5 lines
		"\033[2K", // Clear entire line
		"\033[1E", // Move to next line
		"\033[5F", // Move back up 5 lines
	}
	for _, seq := range expectedSequences {
		if !strings.Contains(output, seq) {
			t.Errorf("Clear didn't output expected sequence: %q", seq)
		}
	}

	// Check that it doesn't contain the full screen clear sequence
	if strings.Contains(output, "\033[J") {
		t.Errorf("Clear shouldn't have cleared the full screen")
	}
}

func TestClearPartial(t *testing.T) {
	t.Parallel()

	w := &mockWriter{}
	sb := NewScreenBuf(w)
	sb.height = 5 // Set a fixed height for testing

	// Write fewer lines than the buffer height
	lines := []string{"1", "2", "3"}
	for _, line := range lines {
		sb.WriteLine(line)
	}

	sb.Clear()

	if len(sb.lines) != 0 {
		t.Errorf("Expected lines to be empty after clear, got %v", sb.lines)
	}

	output := w.buf.String()
	expectedSequences := []string{
		"\033[3F", // Move up 3 lines
		"\033[2K", // Clear entire line
		"\033[1E", // Move to next line
		"\033[3F", // Move back up 3 lines
	}
	for _, seq := range expectedSequences {
		if !strings.Contains(output, seq) {
			t.Errorf("Clear didn't output expected sequence: %q", seq)
		}
	}

	// Check that it doesn't contain the full screen clear sequence
	if strings.Contains(output, "\033[J") {
		t.Errorf("Clear shouldn't have cleared the full screen")
	}
}

func TestClearEmpty(t *testing.T) {
	t.Parallel()

	w := &mockWriter{}
	sb := NewScreenBuf(w)

	// Clear an empty buffer
	sb.Clear()

	if len(sb.lines) != 0 {
		t.Errorf("Expected lines to be empty after clear, got %v", sb.lines)
	}

	output := w.buf.String()
	if output != "" {
		t.Errorf(
			"Clear on empty buffer shouldn't output anything, got: %q",
			output,
		)
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	sb := &ScreenBuf{
		lines: []string{"Line 1", "Line 2", "Line 3"},
	}

	expected := "Line 1\nLine 2\nLine 3"
	if sb.String() != expected {
		t.Errorf("Expected string %s, got %s", expected, sb.String())
	}
}

func TestScreenBufIntegration(t *testing.T) {
	t.Parallel()

	w := &mockWriter{}
	sb := NewScreenBuf(w)
	sb.height = 5 // Set a fixed height for testing

	// Write more lines than the buffer can hold
	for i := 1; i <= 10; i++ {
		sb.WriteLine(fmt.Sprintf("Line %d", i))
	}

	// Check if only the last 5 lines are kept
	expected := []string{"Line 6", "Line 7", "Line 8", "Line 9", "Line 10"}
	if !equalSlices(sb.lines, expected) {
		t.Errorf("Expected lines %v, got %v", expected, sb.lines)
	}

	// Clear the buffer
	sb.Clear()

	// Write some more lines
	sb.WriteLine("New line 1")
	sb.WriteLine("New line 2")

	// Check if the new lines are there
	expected = []string{"New line 1", "New line 2"}
	if !equalSlices(sb.lines, expected) {
		t.Errorf("Expected lines %v, got %v", expected, sb.lines)
	}

	// Check the final string representation
	expectedStr := "New line 1\nNew line 2"
	if sb.String() != expectedStr {
		t.Errorf("Expected string %s, got %s", expectedStr, sb.String())
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
