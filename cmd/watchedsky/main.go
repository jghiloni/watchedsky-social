package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jghiloni/watchedsky-social/backend/appcontext"
	"github.com/jghiloni/watchedsky-social/backend/config"
	"github.com/jghiloni/watchedsky-social/backend/daemons"
)

func main() {
	rawCtx, err := config.LoadAppConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	loaded, err := appcontext.Registry.LoadClients(rawCtx)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(loaded)
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals

		cancel()
	}()

	go daemons.HTTPServerDaemon(ctx)
	go daemons.NWSAlertPoller(ctx)
	go daemons.FirehoseConsumer(ctx)
}
