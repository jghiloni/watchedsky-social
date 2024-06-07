package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"

	"github.com/jghiloni/watchedsky-social/backend/features"
	"github.com/dsnet/compress/bzip2"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))

	logger.Info("Getting zones")

	resp, err := http.Get("https://api.weather.gov/zones")
	if err != nil {
		logger.Error("error getting zones", slog.Any("err", err))
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("error getting zones", slog.Any("err", fmt.Errorf("expected 200, got %d", resp.StatusCode)))
		os.Exit(1)
	}

	var fc features.FeatureCollection
	if err = json.NewDecoder(resp.Body).Decode(&fc); err != nil {
		logger.Error("error parsing zones", slog.Any("err", err))
		os.Exit(1)
	}

	zones := make([]features.Feature, 0, len(fc.Features))
	for _, feat := range fc.Features {
		logger.Info("getting zone", slog.String("zone id", path.Base(feat.ID)))
		r, err := http.Get(feat.ID)
		if err != nil {
			logger.Error("error getting zone", slog.Any("err", err))
		}
		defer r.Body.Close()

		if r.StatusCode != http.StatusOK {
			logger.Error("error getting zone", slog.Any("err", fmt.Errorf("expected 200, got %d", r.StatusCode)))
			os.Exit(1)
		}

		var f features.Feature
		if err = json.NewDecoder(r.Body).Decode(&f); err != nil {
			logger.Error("error parsing zone", slog.String("zone id", path.Base(feat.ID)))
			continue
		}

		zones = append(zones, f)
	}

	jsonout, err := os.Create("zones.json.bz2")
	if err != nil {
		logger.Error("can't create json output", slog.Any("err", err))
		os.Exit(1)
	}
	defer jsonout.Close()

	jsonWriter, err := bzip2.NewWriter(jsonout, &bzip2.WriterConfig{
		Level: bzip2.BestCompression,
	})
	if err != nil {
		logger.Error("can't create compressor", slog.Any("err", err))
		os.Exit(1)
	}
	defer jsonWriter.Close()

	if err = json.NewEncoder(jsonWriter).Encode(zones); err != nil {
		logger.Error("json failed", slog.Any("err", err))
		os.Exit(1)
	}
}
