package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/jghiloni/watchedsky-social/backend/appcontext"
	"github.com/jghiloni/watchedsky-social/backend/daemons"
)

func main() {
	ctx, cancel := appcontext.BuildApplicationContext()
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals

		cancel()
	}()

	go daemons.HTTPServerDaemon(ctx)
	go daemons.NWSAlertPoller(ctx)
}
