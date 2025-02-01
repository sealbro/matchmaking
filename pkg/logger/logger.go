package logger

import (
	"log/slog"
	"os"
)

func NewLogger(logLevel string) *slog.Logger {
	level := toSlogLevel(logLevel)

	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger
}

func toSlogLevel(level string) slog.Level {
	switch level {
	case slog.LevelDebug.String():
		return slog.LevelDebug
	case slog.LevelInfo.String():
		return slog.LevelInfo
	case slog.LevelWarn.String():
		return slog.LevelWarn
	case slog.LevelError.String():
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
