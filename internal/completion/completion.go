package completion

import "time"

// Completion is a union type that can be either a CompletionData or a CompletionTombStone.
// This is used to represent completions in the completion history.
// A completion will always end up as a CompletionTomStone after it has been used.
type Completion interface {
	Content() string
	Timestamp() time.Time
}
