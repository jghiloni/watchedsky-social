package logging

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/jghiloni/watchedsky-social/backend/appcontext"
	"github.com/jghiloni/watchedsky-social/backend/config"
)

type contextKey struct{}

var loggerContextKey contextKey

func loadClientToContext(ctx context.Context, cfg config.AppConfig) (context.Context, error) {
	level := cfg.LogLevel.SLogLevel()
	var out io.Writer = os.Stdout
	if level == config.SlogOff {
		out = io.Discard
		level = slog.LevelError
	}

	logger := slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	}))

	ctx = context.WithValue(ctx, loggerContextKey, logger)
	return ctx, nil
}

func init() {
	appcontext.Registry.RegisterClient(loadClientToContext)
}

func GetLogger(ctx context.Context) *slog.Logger {
	log, _ := ctx.Value(loggerContextKey).(*slog.Logger)
	return log
}
