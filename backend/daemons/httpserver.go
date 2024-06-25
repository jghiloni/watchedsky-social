package daemons

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/storage/mongodb/v2"
	feedhttp "github.com/jghiloni/go-bsky-feed-generator/http"
	"github.com/jghiloni/watchedsky-social/backend/api"
	"github.com/jghiloni/watchedsky-social/backend/config"
	"github.com/jghiloni/watchedsky-social/frontend"
)

func isProd() bool {
	return os.Getenv("RUNTIME_ENV") == "PROD"
}

// HTTPServerDaemon runs the HTTP APIs for WatchedSky, including the front end,
// REST APIs, and feed generators
func HTTPServerDaemon(ctx context.Context) error {
	cfg := config.GetConfig(ctx)
	port := 10000

	if !cfg.HTTPServer.Enabled {
		return nil
	}

	port = int(cfg.HTTPServer.Port)

	app := fiber.New(fiber.Config{
		AppName:               "WatchedSky",
		CaseSensitive:         true,
		DisableStartupMessage: isProd(),
		ETag:                  true,
		GETOnly:               true,
		Immutable:             false,
		Prefork:               false,
		ServerHeader:          "WatchedSky",
		StreamRequestBody:     true,
	})

	app.Use(
		cacheMiddleware(ctx),
		compress.New(),
		etag.New(),
		favicon.New(),
		filesystem.New(filesystem.Config{
			Root:       http.FS(frontend.BuiltSite),
			PathPrefix: "/dist",
			Browse:     false,
			Index:      "index.html",
		}),
		healthcheckMiddleware(),
		helmet.New(),
		requestid.New(), // must be before logger.New
		logger.New(logger.Config{
			TimeZone:      time.UTC.String(),
			TimeFormat:    time.RFC3339,
			DisableColors: isProd(),
			Format:        `${ip} ${locals:requestid} [${time}] ${latency} ${method} ${path} ${status} ${bytesSent}\n`,
		}),
		recover.New(),
	)

	apiGroup := app.Group("/api")
	features := apiGroup.Group("/features", api.ListFeatures(ctx))
	features.Get("/:id", api.GetFeature(ctx))

	app.Get("/xrpc/app.bsky.feed.getFeedSkeleton", adaptor.HTTPHandler(feedhttp.FeedHandler(ctx, nil)))

	return app.Listen(fmt.Sprintf(":%d", port))
}

func cacheMiddleware(ctx context.Context) fiber.Handler {
	cfg := config.GetConfig(ctx)

	host := ""
	port := 0
	username := ""
	password := ""

	hp := strings.Split(cfg.MongoDB.Host, ":")
	host = hp[0]
	if len(hp) > 1 {
		port, _ = strconv.Atoi(hp[1])
	}

	username = cfg.MongoDB.Username
	password = cfg.MongoDB.Password

	cacheStorage := mongodb.New(mongodb.Config{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Reset:    false,
	})

	return cache.New(cache.Config{
		CacheControl: true,
		Expiration:   time.Hour,
		Storage:      cacheStorage,
		Next: func(c *fiber.Ctx) bool {
			return strings.HasPrefix(c.Path(), "/api/search") || strings.HasPrefix(c.Path(), "/xrpc/")
		},
	})
}

func healthcheckMiddleware() fiber.Handler {
	return healthcheck.New(healthcheck.Config{
		LivenessEndpoint:  "/health",
		ReadinessEndpoint: "/ready",
		ReadinessProbe: func(c *fiber.Ctx) bool {
			return AreDaemonsHealthy()
		},
	})
}
