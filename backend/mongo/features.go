package mongo

import (
	"context"
	"fmt"

	"github.com/jghiloni/watchedsky-social/backend/features"
	"github.com/jghiloni/watchedsky-social/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PageOptions struct {
	Page     uint `json:"page"`
	PageSize uint `json:"pageSize"`
}

type FeaturePage struct {
	PageInfo PageOptions       `json:",inline"`
	Features features.Features `json:"features"`
}

func (c *MongoClient) ListFeaturesByType(ctx context.Context, featureType string, pageInfo PageOptions) (FeaturePage, error) {
	coll := c.cli.Collection("features")

	if pageInfo.PageSize < 1 {
		pageInfo.PageSize = 1
	}

	if pageInfo.PageSize > 500 {
		pageInfo.PageSize = 500
	}

	query := bson.D{}
	if featureType != "" {
		query = bson.D{{"properties.@type", featureType}}
	}

	cursor, err := coll.Find(ctx, query, &options.FindOptions{
		Limit: utils.Ptr(int64(pageInfo.PageSize)),
		Skip:  utils.Ptr(int64(pageInfo.PageSize * pageInfo.Page)),
	})
	if err != nil {
		return FeaturePage{}, err
	}
	defer cursor.Close(ctx)

	feats := make([]features.Feature, 0, pageInfo.PageSize)
	for cursor.Next(ctx) {
		var f features.Feature
		if err = cursor.Decode(&f); err != nil {
			return FeaturePage{}, fmt.Errorf("could not decode feature: %w", err)
		}

		feats = append(feats, f)
	}

	return FeaturePage{
		PageInfo: PageOptions{
			Page:     pageInfo.Page,
			PageSize: utils.WNMin(pageInfo.PageSize, uint(len(feats))),
		},
		Features: feats,
	}, nil
}

func (c *MongoClient) GetFeaturesByID(ctx context.Context, ids ...string) (features.FeatureCollection, error) {
	coll := c.cli.Collection("features")

	query := bson.D{{"_id", bson.D{{"$in", bson.A(utils.AnySlice(ids))}}}}
	cursor, err := coll.Find(ctx, query)
	if err != nil {
		return features.FeatureCollection{}, err
	}
	defer cursor.Close(ctx)

	feats := make([]features.Feature, 0, len(ids))
	for cursor.Next(ctx) {
		var f features.Feature
		if err = cursor.Decode(&f); err != nil {
			return features.FeatureCollection{}, fmt.Errorf("could not decode feature: %w", err)
		}

		feats = append(feats, f)
	}

	return features.FeatureCollection{
		Features: feats,
	}, nil
}

func (c *MongoClient) AddFeatures(ctx context.Context, feats ...features.Feature) error {
	coll := c.cli.Collection("features")
	_, err := coll.InsertMany(ctx, utils.AnySlice(feats), &options.InsertManyOptions{
		Ordered: utils.Ptr(true),
	})

	return err
}
