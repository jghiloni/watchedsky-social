package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jghiloni/watchedsky-social/backend/daemons"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	databaseArgs
	daemons.AlertDaemon
	Port           uint16 `env:"WATCHEDSKY_SERVER_PORT" default:"10000" help:"The port for the server to listen on"`
	MetricsEnabled bool   `env:"WATCHEDSKY_SERVER_METRICS_ENABLED" default:"true" negatable:"" help:"If true, enable Prometheus metrics"`
}

func (s *Server) Run(ctx context.Context, logger *slog.Logger) error {
	if err := s.login(ctx); err != nil {
		return err
	}

	errorQueue := make(chan error, 1)
	s.StartDaemon(&s.AlertDaemon, ctx, logger, false, errorQueue)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ServerHeader:          "WatchedSky",
		CaseSensitive:         true,
		AppName:               "WatchedSky",
		Prefork:               false,
	})

	if s.MetricsEnabled {
		app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	}

	return app.Listen(fmt.Sprintf(":%d", s.Port))
}

func (s *Server) StartDaemon(d daemons.ServerDaemon, ctx context.Context, logger *slog.Logger, stopOnFirstError bool, errorQueue chan<- error) {
	logger.Info("StartDaemon")
	go func() {
		err := d.Start(ctx, logger, s.db, stopOnFirstError, errorQueue)
		logger.DebugContext(ctx, "daemon ended with an error", slog.Any("error", err))
		if !errors.Is(err, daemons.ErrStopped) {
			errorQueue <- err
		}
	}()
}
