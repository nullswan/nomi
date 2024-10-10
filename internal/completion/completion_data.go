package completion

import "time"

// Data is a struct that represents the data of a completion.
type Data struct {
	timestamp time.Time
	content   string
}

func (c Data) Content() string {
	return c.content
}

func (c Data) Timestamp() time.Time {
	return c.timestamp
}

func NewCompletionData(content string) Data {
	return Data{
		timestamp: time.Now(),
		content:   content,
	}
}

func (c Data) WithContent(content string) Data {
	c.content = content
	return c
}

func (c Data) WithTimestamp(t time.Time) Data {
	c.timestamp = t
	return c
}
