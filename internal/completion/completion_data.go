package completion

import "time"

// CompletionData is a struct that represents the data of a completion.
type CompletionData struct {
	timestamp time.Time
	content   string
}

func (c CompletionData) Content() string {
	return c.content
}

func (c CompletionData) Timestamp() time.Time {
	return c.timestamp
}

func NewCompletionData(content string) CompletionData {
	return CompletionData{
		timestamp: time.Now(),
		content:   content,
	}
}

func (c CompletionData) WithContent(content string) CompletionData {
	c.content = content
	return c
}

func (c CompletionData) WithTimestamp(t time.Time) CompletionData {
	c.timestamp = t
	return c
}
