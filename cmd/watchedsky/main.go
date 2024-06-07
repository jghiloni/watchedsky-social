package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jghiloni/watchedsky-social/backend/commands"
	"github.com/alecthomas/kong"
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

type CLIArgs struct {
	Server   *commands.Server `cmd:"" help:"Start the HTTP server, with all background processes"`
	LogLevel LogLevel         `env:"WATCHEDSKY_LOG_LEVEL" enum:"off,error,warn,info,debug" default:"info" help:"Log Level (off, error, warn, info, debug)"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		<-signals

		cancel()
	}()

	var args CLIArgs
	kctx := kong.Parse(&args)

	kctx.Bind(getLogger(kctx, args))
	kctx.BindTo(ctx, (*context.Context)(nil))

	kctx.FatalIfErrorf(kctx.Run())
}

func getLogger(ctx *kong.Context, args CLIArgs) *slog.Logger {
	out := ctx.Stdout
	level := args.LogLevel.SLogLevel()
	if level == slogOff {
		out = io.Discard
		level = slog.LevelError
	}

	return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	}))
}
