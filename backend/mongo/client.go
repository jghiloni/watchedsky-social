package mongo

import (
	"context"
	"fmt"

	"github.com/jghiloni/watchedsky-social/backend/appcontext"
	"github.com/jghiloni/watchedsky-social/backend/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type contextKey struct{}

var clientContextKey contextKey

type MongoClient struct {
	cli *mongo.Database
}

func loadClientToContext(ctx context.Context, cfg config.AppConfig) (context.Context, error) {
	if cfg.MongoDB.Username == "" {
		return ctx, nil
	}

	dbConfig := mergeConfigs(cfg)

	dsn := fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s",
		dbConfig.Username, dbConfig.Password, dbConfig.Host,
		dbConfig.Name, dbConfig.AuthenticationDatabase,
	)

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(dsn).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	dbClient := client.Database(dbConfig.Name, options.Database().SetBSONOptions(&options.BSONOptions{
		UseJSONStructTags:       true,
		ErrorOnInlineDuplicates: false,
		NilMapAsEmpty:           true,
		NilSliceAsEmpty:         true,
		NilByteSliceAsEmpty:     true,
	}))

	ctx = context.WithValue(ctx, clientContextKey, &MongoClient{cli: dbClient})
	return ctx, nil
}

func mergeConfigs(cfg config.AppConfig) config.DatabaseConfig {
	retConfig := config.DatabaseConfig{
		Host:                   "localhost:27017",
		Name:                   "watchedsky",
		AuthenticationDatabase: "admin",
	}

	mongoCfg := cfg.MongoDB

	if mongoCfg.Host != "" {
		retConfig.Host = cfg.MongoDB.Host
	}

	if mongoCfg.Username != "" {
		retConfig.Username = cfg.MongoDB.Username
	}

	if mongoCfg.Password != "" {
		retConfig.Password = cfg.MongoDB.Password
	}

	return retConfig
}

func init() {
	appcontext.Registry.RegisterClient(loadClientToContext)
}

func GetClient(ctx context.Context) *MongoClient {
	cli, _ := ctx.Value(clientContextKey).(*MongoClient)
	return cli
}
