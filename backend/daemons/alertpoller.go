package daemons

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jghiloni/watchedsky-social/backend/appcontext"
	"github.com/jghiloni/watchedsky-social/backend/atproto"
	"github.com/jghiloni/watchedsky-social/backend/features"
)

const apiURL = "https://api.weather.gov/alerts/active?status=actual&urgency=Immediate,Expected,Future,Unknown&certainty=Observed,Likely,Possible,Unknown"

// NWSAlertPoller polls the NWS API for new alerts and, when found, pushes them
// to a PDS, which will then be pulled via the firehose and stored in the DB and
// published as skeets
func NWSAlertPoller(ctx context.Context) {
	cfg := appcontext.AppConfig(ctx)
	if cfg == nil {
		panic("configuration could not be found!")
	}

	if !cfg.AlertPoller.Enabled {
		return
	}

	bskyClient, err := atproto.NewBlueskyClient(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		// i know this is bass-ackwards, but i don't want to fail the daemon on
		// a single failed request, and it seems cleaner to put the error handling
		// at the bottom in this case. whatever, it's my code. you're not my real
		// dad!
		latestID, err := bskyClient.GetLatestID(ctx)
		if err == nil {

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
			if err != nil {
				return
			}

			req.Header.Set("User-Agent", "watchedsky.social/monitor")
			req.Header.Set("Accept", "application/geo+json")

			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					var fs features.FeatureCollection
					if err = json.NewDecoder(resp.Body).Decode(&fs); err == nil {
						for _, feat := range fs.Features {
							if feat.Properties.StringValue("id") == latestID {
								break
							}

							if err = bskyClient.PostAlert(ctx, feat); err != nil {
								log.Println(err)
							}
						}
					} else {
						log.Println(err)
					}
				} else {
					log.Printf("expected status code 200 from NWS API, got %d", resp.StatusCode)
				}
			} else {
				log.Println(err)
			}
		} else {
			log.Println(err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(cfg.AlertPoller.PollInterval):
			continue
		}
	}
}
