package daemons

import (
	"context"
	"log/slog"
	"sync"

	"github.com/jghiloni/watchedsky-social/backend/appcontext"
	"github.com/jghiloni/watchedsky-social/backend/config"
	"github.com/jghiloni/watchedsky-social/backend/logging"
)

// Daemon is a func that should be run as goroutines
type Daemon func(context.Context) error

var enabledDaemons map[string]bool
var runningDaemons map[string]bool = map[string]bool{}

var runMu *sync.Mutex = new(sync.Mutex)

func setEnabledDaemons(ctx context.Context, cfg config.AppConfig) (context.Context, error) {
	enabledDaemons = map[string]bool{
		"HTTPServer":     cfg.HTTPServer.Enabled,
		"AlertPoller":    cfg.AlertPoller.Enabled,
		"FirehoseNozzle": cfg.FirehoseNozzle.Enabled,
	}

	return ctx, nil
}

func init() {
	appcontext.Registry.RegisterClient(setEnabledDaemons)
}

func StartDaemon(ctx context.Context, name string, daemon Daemon) {
	runMu.Lock()
	runningDaemons[name] = true
	runMu.Unlock()

	defer func() {
		runMu.Lock()
		runningDaemons[name] = false
		runMu.Unlock()
	}()

	logger := logging.GetLogger(ctx)
	if err := daemon(ctx); err != nil {
		logger.Error("error running daemon", slog.String("daemon", name), slog.Any("err", err))
	}
}

func AreDaemonsHealthy() bool {
	for d, e := range enabledDaemons {
		if e {
			if !runningDaemons[d] {
				return false
			}
		}
	}

	return true
}
