package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jghiloni/watchedsky-social/backend/mongo"
)

func ListFeatures(ctx context.Context) fiber.Handler {
	return func(c *fiber.Ctx) error {
		mongoClient := mongo.GetClient(ctx)
		if mongoClient == nil {
			return errors.New("no mongo client configured")
		}

		pageSize := c.QueryInt("limit", 100)
		if pageSize < 1 {
			pageSize = 1
		}

		if pageSize > 500 {
			pageSize = 500
		}

		page := c.QueryInt("page", 0)
		if page < 0 {
			page = 0
		}
		featureType := c.Query("type")

		response, err := mongoClient.ListFeaturesByType(ctx, featureType, mongo.PageOptions{
			Page:     uint(page),
			PageSize: uint(pageSize),
		})

		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
		}

		return c.JSON(response)
	}
}

func GetFeature(ctx context.Context) fiber.Handler {
	return func(c *fiber.Ctx) error {
		mongoClient := mongo.GetClient(ctx)
		if mongoClient == nil {
			return errors.New("no mongo client configured")
		}

		featureID := c.Params("id")

		f, err := mongoClient.GetFeaturesByID(ctx, featureID)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
		}

		return c.JSON(f)
	}
}
