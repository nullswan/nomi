package tools

import "fmt"

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Error(msg string)
	Println(msg string)
}

type logger struct {
	devMode bool
}

func NewLogger(
	devMode bool,
) Logger {
	return &logger{
		devMode: devMode,
	}
}

func (l *logger) Debug(msg string) {
	if !l.devMode {
		return
	}

	fmt.Println("[Debug] " + msg)
}

func (l *logger) Info(msg string) {
	fmt.Println("[Info] " + msg)
}

func (l *logger) Error(msg string) {
	fmt.Println("[Error] " + msg)
}

func (l *logger) Println(msg string) {
	fmt.Println(msg)
}
