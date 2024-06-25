package daemons

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	"github.com/bluesky-social/indigo/repo"
	"github.com/gorilla/websocket"
	"github.com/ipfs/go-cid"
	"github.com/jghiloni/watchedsky-social/backend/config"
	"github.com/jghiloni/watchedsky-social/backend/mongo"

	"github.com/jghiloni/watchedsky-social/backend/bsky"
)

const firehoseURI = "https://bsky.network/xrpc/com.atproto.sync.subscribeRepos"

func FirehoseConsumer(ctx context.Context) {
	cfg := config.GetConfig(ctx)
	if !cfg.FirehoseNozzle.Enabled {
		return
	}

	bskyClient := bsky.GetClient(ctx)
	if bskyClient == nil {
		return
	}

	dbClient := mongo.GetClient(ctx)
	if dbClient == nil {
		return
	}

	me := bskyClient.Me()

	con, _, err := websocket.DefaultDialer.Dial(firehoseURI, http.Header{})
	if err != nil {
		log.Println(err)
		return
	}

	callbacks := &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			if evt.Repo == me.Handle || evt.Repo == me.Did {
				for _, op := range evt.Ops {
					if op.Action == "create" {
						if strings.HasPrefix(op.Path, bsky.AlertCollection+"/") {
							r, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(evt.Blocks))
							if err != nil {
								return err
							}

							rc, rec, err := r.GetRecord(ctx, op.Path)
							if err != nil {
								return err
							}

							if rc != cid.Cid(*op.Cid) {
								return errors.New("cid mismatch. wat?")
							}

							alert, ok := rec.(*bsky.Alert)
							if !ok {
								return fmt.Errorf("why is rec a %T and not an alert?", rec)
							}

							feat, err := alert.ToFeature(ctx)
							if err != nil {
								return err
							}

							if err = dbClient.AddFeatures(ctx, feat); err != nil {
								return err
							}

							return bskyClient.SkeetAlert(ctx, alert)
						}
					}
				}
			}

			return nil
		},
	}

	scheduler := sequential.NewScheduler("watchedsky.social", callbacks.EventHandler)
	events.HandleRepoStream(ctx, con, scheduler)
}
