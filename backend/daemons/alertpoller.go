package daemons

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jghiloni/watchedsky-social/backend/bsky"
	"github.com/jghiloni/watchedsky-social/backend/config"
	"github.com/jghiloni/watchedsky-social/backend/features"
	"github.com/jghiloni/watchedsky-social/backend/logging"
)

const apiURL = "https://api.weather.gov/alerts/active?status=actual&urgency=Immediate,Expected,Future,Unknown&certainty=Observed,Likely,Possible,Unknown"

// AlertPoller polls the NWS API for new alerts and, when found, pushes them
// to a PDS, which will then be pulled via the firehose and stored in the DB and
// published as skeets
func AlertPoller(ctx context.Context) error {
	cfg := config.GetConfig(ctx)
	logger := logging.GetLogger(ctx)

	if !cfg.AlertPoller.Enabled {
		return nil
	}

	logger.Info("starting alert poller")

	bskyClient := bsky.GetClient(ctx)
	if bskyClient != nil {
		logger.Warn("bluesky config not set, exiting")
		return nil
	}

	for {
		// i know this is bass-ackwards, but i don't want to fail the daemon on
		// a single failed request, and it seems cleaner to put the error handling
		// at the bottom in this case. whatever, it's my code. you're not my real
		// dad!
		latestID, err := bskyClient.GetLatestID(ctx)
		if err == nil {
			logger.Info("getting NWS alerts since", slog.String("latestID", latestID))
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
			if err != nil {
				return err
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
								logger.Error("error posting alert to PDS", slog.Any("err", err))
							}
						}
					} else {
						logger.Error("error parsing feature geojson", slog.Any("err", err))
					}
				} else {
					logger.Error(fmt.Sprintf("expected status code 200 from NWS API, got %d", resp.StatusCode))
				}
			} else {
				logger.Error("error getting geojson from NWS API", slog.Any("err", err))
			}
		} else {
			logger.Error("couldn't get latest ID from PDS", slog.Any("err", err))
		}

		select {
		case <-ctx.Done():
			logger.Info("context done, exiting")
			return nil
		case <-time.After(cfg.AlertPoller.PollInterval):
			continue
		}
	}
}
