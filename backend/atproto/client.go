package atproto

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bluesky-social/indigo/api/atproto"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/jghiloni/watchedsky-social/backend/appcontext"
	"github.com/jghiloni/watchedsky-social/backend/features"
	"github.com/jghiloni/watchedsky-social/backend/utils"
)

type BlueskyClient struct {
	xc *xrpc.Client
}

func NewBlueskyClient(ctx context.Context) (*BlueskyClient, error) {
	cfg := appcontext.AppConfig(ctx)
	xrpcClient := &xrpc.Client{
		Client:    util.RobustHTTPClient(),
		Host:      cfg.Bluesky.PDSURL,
		UserAgent: utils.Ptr("watchedsky-social/xrpc-client"),
	}

	loginInput := &atproto.ServerCreateSession_Input{
		Identifier: cfg.Bluesky.Username,
		Password:   cfg.Bluesky.AppPassword,
	}

	authResult, err := atproto.ServerCreateSession(ctx, xrpcClient, loginInput)
	if err != nil {
		return nil, fmt.Errorf("could not log in to Bluesky: %w", err)
	}

	xrpcClient.Auth = &xrpc.AuthInfo{
		AccessJwt:  authResult.AccessJwt,
		RefreshJwt: authResult.RefreshJwt,
		Handle:     authResult.Handle,
		Did:        authResult.Did,
	}

	return &BlueskyClient{xc: xrpcClient}, nil
}

func (c *BlueskyClient) PostAlert(ctx context.Context, f features.Feature) error {
	// first, if the feature is not an alert, return an error
	if f.Properties.StringValue("@type") != features.Alert {
		return errors.New("only Alert features are supported")
	}

	alert := FromFeature(f)

	// if the geometry is not nil, marshal it to json and upload it as a blob
	if f.Geometry != nil {
		jsonBytes, err := json.Marshal(f.Geometry)
		if err != nil {
			return fmt.Errorf("error serializing alert geojson: %w", err)
		}

		blobOutput, err := atproto.RepoUploadBlob(ctx, c.xc, bytes.NewBuffer(jsonBytes))
		if err != nil {
			return fmt.Errorf("error uploading blob: %w", err)
		}

		alert.Geometry = blobOutput.Blob
	}

	_, err := atproto.RepoCreateRecord(ctx, c.xc, &atproto.RepoCreateRecord_Input{
		Collection: "social.watchedsky.alert",
		Repo:       c.xc.Auth.Handle,
		Record: &lexutil.LexiconTypeDecoder{
			Val: &alert,
		},
	})

	return err
}

func (c *BlueskyClient) GetLatestID(ctx context.Context) (string, error) {
	out, err := atproto.RepoListRecords(ctx, c.xc, "social.watchedsky.alert", "", 1, c.xc.Auth.Handle, false, "", "")
	if err != nil {
		return "", err
	}

	alertVal := out.Records[0].Value.Val
	alert, ok := alertVal.(*Alert)
	if ok {
		return alert.Id, nil
	}

	return "", fmt.Errorf("expected record to be an *atproto.Alert, but it was %T", alertVal)
}
