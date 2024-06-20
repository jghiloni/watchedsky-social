package appcontext

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/bluesky-social/indigo/atproto/syntax"
	feedconfig "github.com/jghiloni/go-bsky-feed-generator/config"
	"github.com/jghiloni/watchedsky-social/backend/config"
	"github.com/jghiloni/watchedsky-social/backend/logging"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type contextKey string

const (
	appConfig   = contextKey("--ws-config--")
	dbClientKey = contextKey("--ws-db-client--")
	rmqConnKey  = contextKey("--ws-rmq-conn--")
	loggerKey   = contextKey("--ws-log-key--")
)

func BuildApplicationContext() (context.Context, context.CancelFunc) {
	cfg, err := config.LoadAppConfig()
	if err != nil {
		panic(err)
	}

	ctx := context.WithValue(context.Background(), appConfig, cfg)

	dsn := fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s",
		cfg.MongoDB.Username, cfg.MongoDB.Password, cfg.MongoDB.Host,
		cfg.MongoDB.Name, cfg.MongoDB.AuthenticationDatabase,
	)

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(dsn).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}

	dbClient := client.Database(cfg.MongoDB.Name, options.Database().SetBSONOptions(&options.BSONOptions{
		UseJSONStructTags:       true,
		ErrorOnInlineDuplicates: false,
		NilMapAsEmpty:           true,
		NilSliceAsEmpty:         true,
		NilByteSliceAsEmpty:     true,
	}))

	ctx = context.WithValue(ctx, dbClientKey, dbClient)

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s", cfg.RabbitMQ.Username, cfg.RabbitMQ.Password, cfg.RabbitMQ.Host))
	if err != nil {
		panic(err)
	}

	ctx = context.WithValue(ctx, rmqConnKey, conn)
	ctx = context.WithValue(ctx, loggerKey, logging.GetLogger(os.Stdout, cfg.LogLevel))

	ctx = feedconfig.WithConfig(ctx, feedconfig.FeedGeneratorConfig{
		ServiceDID: syntax.ATURI(cfg.Bluesky.FeedServiceDID),
	})

	return context.WithCancel(ctx)
}

func AppConfig(ctx context.Context) *config.AppConfig {
	if ctx == nil {
		return nil
	}

	appCfg, ok := ctx.Value(appConfig).(*config.AppConfig)
	if !ok {
		return nil
	}

	return appCfg
}

func DBClient(ctx context.Context) *mongo.Database {
	if ctx == nil {
		return nil
	}

	dbClient, ok := ctx.Value(dbClientKey).(*mongo.Database)
	if !ok {
		return nil
	}

	return dbClient
}

func RabbitMQConnection(ctx context.Context) *amqp.Connection {
	if ctx == nil {
		return nil
	}

	rmqConn, ok := ctx.Value(rmqConnKey).(*amqp.Connection)
	if !ok {
		return nil
	}

	return rmqConn
}

func Logger(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return nil
	}

	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok {
		return nil
	}

	return logger
}
