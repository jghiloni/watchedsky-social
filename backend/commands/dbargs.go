package commands

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type databaseArgs struct {
	Username               string          `required:"" xor:"dbuser" env:"WATCHEDSKY_DB_USERNAME" help:"The MongoDB username"`
	Password               string          `required:"" xor:"dbpass" env:"WATCHEDSKY_DB_PASSWORD" help:"The MongoDB password"`
	Hostname               string          `required:"" xor:"dbhost" env:"WATCHEDSKY_DB_HOSTNAME" help:"The MongoDB hostname"`
	DatabasePort           uint16          `required:"" xor:"dbport" env:"WATCHEDSKY_DB_PORT" help:"The MongoDB port"`
	Database               string          `required:"" xor:"dbname" env:"WATCHEDSKY_DB_NAME" help:"The MongoDB database name"`
	AuthenticationDatabase string          `required:"" xor:"dbauth" env:"WATCHEDSKY_DB_AUTHDB" help:"The MongoDB authentication database"`
	DSN                    string          `required:"" xor:"dbuser,dbpass,dbhost,dbport,dbname,dbauth" env:"WATCHEDSKY_DB_DSN" help:"The MongoDB DSN. Optional if all other db parameters are set"`
	db                     *mongo.Database `kong:"-"`
}

func (d databaseArgs) GetDSN() string {
	if d.DSN != "" {
		return d.DSN
	}

	return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=%s",
		d.Username, d.Password, d.Hostname, d.DatabasePort, d.Database, d.AuthenticationDatabase)
}

func (d *databaseArgs) login(ctx context.Context) error {
	dsn := d.GetDSN()
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(dsn).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return err
	}

	db := d.Database
	if db == "" {
		u, err := url.Parse(dsn)
		if err != nil {
			return err
		}

		db = strings.TrimPrefix(u.Path, "/")
	}

	d.db = client.Database(db, options.Database().SetBSONOptions(&options.BSONOptions{
		UseJSONStructTags:       true,
		ErrorOnInlineDuplicates: false,
		NilMapAsEmpty:           true,
		NilSliceAsEmpty:         true,
		NilByteSliceAsEmpty:     true,
	}))

	return nil
}
