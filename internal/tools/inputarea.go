package tools

type InputArea interface {
	Read(defaultValue string) string
}

func NewInputArea() InputArea {
	return &inputArea{}
}

type inputArea struct{}

func (i *inputArea) Read(defaultValue string) string {
	return ""
}
