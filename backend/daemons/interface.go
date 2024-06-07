package daemons

import (
	"context"
	"errors"
	"log/slog"

	"go.mongodb.org/mongo-driver/mongo"
)

var ErrStopped = errors.New("daemon stopped")

type ServerDaemon interface {
	Start(ctx context.Context, logger *slog.Logger, db *mongo.Database, stopOnFirstError bool, errors chan<- error) error
}
