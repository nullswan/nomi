package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	logger *slog.Logger
	once   sync.Once
)

func initInstance() {
	level := slog.LevelInfo
	if os.Getenv("DEBUG") != "" {
		level = slog.LevelDebug
	}

	loggerHandlerOpts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	logger = slog.New(
		slog.NewTextHandler(os.Stdout, loggerHandlerOpts),
	)
}

func Init() *slog.Logger {
	once.Do(func() {
		initInstance()
	})

	return logger
}
