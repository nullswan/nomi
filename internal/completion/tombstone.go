package completion

import "time"

// A completion tombstone is the final state of a completion.
type CompletionTombStone struct {
	content   string
	model     string
	usage     Usage
	timestamp time.Time
}

func NewCompletionTombStone(
	content, model string,
	usage Usage,
) CompletionTombStone {
	return CompletionTombStone{
		content:   content,
		model:     model,
		usage:     usage,
		timestamp: time.Now(),
	}
}

func (c CompletionTombStone) Content() string {
	return c.content
}

func (c CompletionTombStone) WithContent(content string) CompletionTombStone {
	c.content = content
	return c
}

func (c CompletionTombStone) Model() string {
	return c.model
}

func (c CompletionTombStone) WithModel(model string) CompletionTombStone {
	c.model = model
	return c
}

func (c CompletionTombStone) Usage() Usage {
	return c.usage
}

func (c CompletionTombStone) WithUsage(usage Usage) CompletionTombStone {
	c.usage = usage
	return c
}

func (c CompletionTombStone) Timestamp() time.Time {
	return c.timestamp
}

func (c CompletionTombStone) WithTimestamp(t time.Time) CompletionTombStone {
	c.timestamp = t
	return c
}
