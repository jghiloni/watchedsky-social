package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jghiloni/watchedsky-social/backend/appcontext"
	"github.com/jghiloni/watchedsky-social/backend/features"
	"github.com/jghiloni/watchedsky-social/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ListFeatures(ctx context.Context) fiber.Handler {
	return func(c *fiber.Ctx) error {
		coll, err := getFeatureCollection(ctx, c)
		if err != nil {
			return err
		}

		pageSize := c.QueryInt("limit", 100)
		if pageSize < 1 {
			pageSize = 1
		}

		if pageSize > 500 {
			pageSize = 500
		}

		page := c.QueryInt("page", 0)
		featureType := c.Query("type")

		query := bson.D{}
		if featureType != "" {
			query = bson.D{{"properties.@type", featureType}}
		}

		cursor, err := coll.Find(ctx, query, &options.FindOptions{
			Limit: utils.Ptr(int64(pageSize)),
			Skip:  utils.Ptr(int64(pageSize * page)),
		})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
		}
		defer cursor.Close(ctx)

		feats := make([]features.Feature, 0, pageSize)
		for cursor.Next(ctx) {
			var f features.Feature
			if err = cursor.Decode(&f); err != nil {
				return fmt.Errorf("could not decode feature: %w", err)
			}

			feats = append(feats, f)
		}

		response := make(map[string]any)
		response["page"] = page
		response["pageSize"] = pageSize
		response["features"] = feats

		return c.JSON(response)
	}
}

func GetFeature(ctx context.Context) fiber.Handler {
	return func(c *fiber.Ctx) error {
		coll, err := getFeatureCollection(ctx, c)
		if err != nil {
			return err
		}

		featureID := c.Params("id")

		query := bson.D{{"_id", featureID}}
		result := coll.FindOne(ctx, query)
		if err := result.Err(); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
		}

		var f features.Feature
		if err := result.Decode(&f); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
		}

		return c.JSON(f)
	}
}

func getFeatureCollection(ctx context.Context, c *fiber.Ctx) (*mongo.Collection, error) {
	db := appcontext.DBClient(ctx)
	if db == nil {
		return nil, c.SendStatus(http.StatusPreconditionFailed)
	}

	return db.Collection("features"), nil
}
