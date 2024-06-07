package daemons

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"time"

	"github.com/jghiloni/watchedsky-social/backend/features"
	"github.com/jghiloni/watchedsky-social/backend/geojson"
	"github.com/jghiloni/watchedsky-social/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AlertDaemon struct {
	AlertAPIPollFrequency time.Duration   `env:"WATCHEDSKY_ALERT_API_POLL_TIME" default:"5m" help:"How frequently to poll for new weather alerts"`
	latestAlertTime       time.Time       `kong:"-"`
	db                    *mongo.Database `kong:"-"`
}

const alertsURL = "https://api.weather.gov/alerts/active?status=actual&message_type=alert&urgency=Immediate,Expected,Future,Unknown&certainty=Observed,Likely,Possible,Unknown"

func (a *AlertDaemon) Start(ctx context.Context, logger *slog.Logger, db *mongo.Database, stopOnFirstError bool, errors chan<- error) error {
	a.db = db
	a.loadLatestAlert(ctx, logger)
	for {
		resp, err := http.Get(alertsURL)
		if err != nil {
			errors <- fmt.Errorf("error fetching alerts: %w", err)
			if stopOnFirstError {
				return err
			}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err := fmt.Errorf("expected 200 OK status, but got %s", resp.Status)
			errors <- err
			if stopOnFirstError {
				return err
			}
		}

		var fc features.FeatureCollection
		if err = json.NewDecoder(resp.Body).Decode(&fc); err != nil {
			errors <- fmt.Errorf("error parsing response: %w", err)
			if stopOnFirstError {
				return err
			}
		}

		sort.Sort(fc.Features)
		alerts := utils.SubsliceUntil(fc.Features, func(alert features.Feature) bool {
			timeStr := alert.Properties["sent"].(string)
			t, err := time.Parse(time.RFC3339, timeStr)
			if err != nil {
				return false
			}

			return t.Before(a.latestAlertTime)
		})

		alertAnys := utils.Map(alerts, func(alert features.Feature) any {
			if alert.Geometry != nil {
				return alert
			}

			gc := &geojson.GeometryCollection{GT: geojson.GeometryCollectionType, Geometries: []geojson.Geometry{}}

			filter := bson.D{{"_id", bson.D{{"$in", alert.Properties["affectedZones"]}}}, {"properties.@type", features.Zone}}
			cursor, err := db.Collection(features.CollectionName).Find(ctx, filter, options.Find().SetAllowPartialResults(false))
			if err != nil {
				logger.ErrorContext(ctx, "error querying mongo", slog.Any("err", err))
				return alert
			}
			defer cursor.Close(ctx)

			for cursor.Next(ctx) {
				var affectedZone features.Feature
				if err := cursor.Decode(&affectedZone); err != nil {
					logger.ErrorContext(ctx, "error decoding zone", slog.Any("err", err))
					return alert
				}

				gc.Geometries = append(gc.Geometries, affectedZone.Geometry)
			}

			if len(gc.Geometries) > 0 {
				alert.Geometry = gc
			}

			return alert
		})

		_, err = db.Collection(features.CollectionName).InsertMany(ctx, utils.Reverse(alertAnys), options.InsertMany().SetOrdered(true))
		if err != nil {
			errors <- err
			if stopOnFirstError {
				return err
			}
		}

		timeStr, ok := alerts[0].Properties["sent"].(string)
		if !ok {
			errors <- fmt.Errorf("sent time not set")
			if stopOnFirstError {
				return err
			}
		}

		a.latestAlertTime, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			errors <- err
			if stopOnFirstError {
				return err
			}
		}

		next := time.Now().Add(a.AlertAPIPollFrequency)
		logger.InfoContext(ctx, "alert check complete",
			slog.Duration("timeTillNextCheck", a.AlertAPIPollFrequency), slog.Time("nextCheckTime", next))

		select {
		case <-time.After(a.AlertAPIPollFrequency):
			continue
		case <-ctx.Done():
			return fmt.Errorf("alert %w", ErrStopped)
		}
	}
}

func (a *AlertDaemon) loadLatestAlert(ctx context.Context, logger *slog.Logger) {
	cursor, err := a.db.Collection(features.CollectionName).Aggregate(ctx, bson.A{
		bson.D{{"$match", bson.D{{"properties.@type", "wx:Alert"}}}},
		bson.D{{"$sort", bson.D{{"properties.sent", -1}}}},
		bson.D{{"$limit", 1}},
	})
	if err != nil {
		logger.ErrorContext(ctx, "error getting latest alert", slog.Any("err", err.Error()))
		return
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var alert features.Feature
		if err = cursor.Decode(&alert); err != nil {
			logger.ErrorContext(ctx, "error decoding latest alert", slog.Any("err", err))
			return
		}

		timeStr, ok := alert.Properties["sent"].(string)
		if !ok {
			logger.ErrorContext(ctx, "sent time not a valid timestamp")
			return
		}

		a.latestAlertTime, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			logger.ErrorContext(ctx, "can't parse timestamp", slog.Any("err", err))
			return
		}
	}
}
