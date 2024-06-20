package logging

import (
	"io"
	"log/slog"
)

type LogLevel string

const (
	Off   LogLevel = "off"
	Error LogLevel = "error"
	Warn  LogLevel = "warn"
	Info  LogLevel = "info"
	Debug LogLevel = "debug"
)

const slogOff slog.Level = slog.Level(-999)

var levelMap map[LogLevel]slog.Level = map[LogLevel]slog.Level{
	Off:   slogOff,
	Error: slog.LevelError,
	Warn:  slog.LevelWarn,
	Info:  slog.LevelInfo,
	Debug: slog.LevelDebug,
}

func (l LogLevel) SLogLevel() slog.Level {
	level, ok := levelMap[l]
	if !ok {
		return slog.LevelInfo
	}

	return level
}

func GetLogger(out io.Writer, logLevel LogLevel) *slog.Logger {
	level := logLevel.SLogLevel()
	if level == slogOff {
		out = io.Discard
		level = slog.LevelError
	}

	return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	}))
}
