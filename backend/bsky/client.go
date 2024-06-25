package bsky

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/jghiloni/watchedsky-social/backend/appcontext"
	"github.com/jghiloni/watchedsky-social/backend/config"
	"github.com/jghiloni/watchedsky-social/backend/features"
	"github.com/jghiloni/watchedsky-social/backend/geojson"
	"github.com/jghiloni/watchedsky-social/backend/utils"
)

const AlertCollection = "social.watchedsky.alert"

type BlueskyClientConfig struct {
	PDSURL   string
	Username string
	Password string
}

var defaultBlueskyConfig = BlueskyClientConfig{
	PDSURL: "https://bsky.network",
}

type BlueskyClient struct {
	xc *xrpc.Client
}

type contextKey struct{}

var clientContextKey contextKey

func loadClientToContext(ctx context.Context, cfg config.AppConfig) (context.Context, error) {
	if cfg.Bluesky.Username == "" {
		return ctx, nil
	}

	client, err := newBlueskyClient(ctx, BlueskyClientConfig{
		PDSURL:   cfg.Bluesky.PDSURL,
		Username: cfg.Bluesky.Username,
		Password: cfg.Bluesky.AppPassword,
	})

	if err != nil {
		return ctx, err
	}

	ctx = context.WithValue(ctx, clientContextKey, client)
	return ctx, nil
}

func newBlueskyClient(ctx context.Context, configs ...BlueskyClientConfig) (*BlueskyClient, error) {
	cfg := mergeConfigs(configs)
	xrpcClient := &xrpc.Client{
		Client:    util.RobustHTTPClient(),
		Host:      cfg.PDSURL,
		UserAgent: utils.Ptr("watchedsky-social/xrpc-client"),
	}

	loginInput := &atproto.ServerCreateSession_Input{
		Identifier: cfg.Username,
		Password:   cfg.Password,
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

func init() {
	appcontext.Registry.RegisterClient(loadClientToContext)
}

func GetClient(ctx context.Context) *BlueskyClient {
	cli, _ := ctx.Value(clientContextKey).(*BlueskyClient)
	return cli
}

func (c *BlueskyClient) Me() *xrpc.AuthInfo {
	if c.xc != nil {
		return c.xc.Auth
	}

	return nil
}

func (c *BlueskyClient) PostAlert(ctx context.Context, f features.Feature) error {
	// first, if the feature is not an alert, return an error
	if f.Properties.StringValue("@type") != features.Alert {
		return errors.New("only Alert features are supported")
	}

	// next, make sure we're logged in
	me := c.Me()
	if me == nil {
		return errors.New("requires auth")
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
		Collection: AlertCollection,
		Repo:       me.Handle,
		Record: &lexutil.LexiconTypeDecoder{
			Val: &alert,
		},
	})

	return err
}

func (c *BlueskyClient) GetAlertGeometry(ctx context.Context, alert Alert) (geojson.Geometry, error) {
	if alert.Geometry == nil {
		return nil, nil
	}

	me := c.Me()
	if me == nil {
		return nil, errors.New("requires auth")
	}

	geoBytes, err := atproto.SyncGetBlob(ctx, c.xc, alert.Geometry.Ref.String(), me.Did)
	if err != nil {
		return nil, err
	}

	var geo geojson.Geometry
	err = json.Unmarshal(geoBytes, &geo)

	return geo, err
}

func (c *BlueskyClient) GetLatestID(ctx context.Context) (string, error) {
	me := c.Me()
	if me == nil {
		return "", errors.New("requires auth")
	}

	out, err := atproto.RepoListRecords(ctx, c.xc, "social.watchedsky.alert", "", 1, me.Handle, false, "", "")
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

func (c *BlueskyClient) SkeetAlert(ctx context.Context, a *Alert) error {
	me := c.Me()
	if me == nil {
		return errors.New("requires auth")
	}

	cfg := config.GetConfig(ctx)

	webURL := fmt.Sprintf("%s/alert/%s", cfg.BaseURL, a.Id)
	msg := fmt.Sprintf("%s Weather Alert: %s. See more at %s", strings.ToUpper(a.Severity), a.Headline, webURL)
	post := bsky.FeedPost{
		CreatedAt: time.Now().Format(time.RFC3339),
		Text:      msg,
	}

	// The last word is a link
	startIdx := strings.Index(msg, webURL)
	if startIdx >= 0 {
		post.Facets = []*bsky.RichtextFacet{
			{
				Index: &bsky.RichtextFacet_ByteSlice{
					ByteStart: int64(startIdx),
					ByteEnd:   int64(len(msg) - 1),
				},
				Features: []*bsky.RichtextFacet_Features_Elem{
					{
						RichtextFacet_Link: &bsky.RichtextFacet_Link{
							Uri: webURL,
						},
					},
				},
			},
		}
	}

	_, err := atproto.RepoCreateRecord(ctx, c.xc, &atproto.RepoCreateRecord_Input{
		Collection: "app.bsky.feed.post",
		Repo:       me.Did,
		Record: &lexutil.LexiconTypeDecoder{
			Val: &post,
		},
	})

	return err
}

func mergeConfigs(configs []BlueskyClientConfig) BlueskyClientConfig {
	config := BlueskyClientConfig{
		PDSURL:   defaultBlueskyConfig.PDSURL,
		Username: defaultBlueskyConfig.Username,
		Password: defaultBlueskyConfig.Password,
	}

	for _, cfg := range configs {
		if cfg.PDSURL != "" {
			config.PDSURL = cfg.PDSURL
		}

		if cfg.Username != "" {
			config.Username = cfg.Username
		}

		if cfg.Password != "" {
			config.Password = cfg.Password
		}
	}

	return config
}
