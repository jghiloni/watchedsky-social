package appcontext

// type appConfigContextKey struct{}
// type dbClientContextKey struct{}
// type bskyClientContextKey struct{}
// type loggerContextKey struct{}

// var (
// 	appConfig     appConfigContextKey
// 	dbClientKey   dbClientContextKey
// 	bskyClientKey bskyClientContextKey
// 	loggerKey     loggerContextKey
// )

// func BuildApplicationContext() (context.Context, context.CancelFunc, error) {
// 	cfg, err := config.LoadAppConfig()
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	// ctx := context.WithValue(context.Background(), appConfig, cfg)
// 	ctx, err := Registry.LoadClients(context.Background())

// 	// if cfg.DBLoader.Enabled {
// 	// 	dsn := fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s",
// 	// 		cfg.MongoDB.Username, cfg.MongoDB.Password, cfg.MongoDB.Host,
// 	// 		cfg.MongoDB.Name, cfg.MongoDB.AuthenticationDatabase,
// 	// 	)

// 	// 	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
// 	// 	opts := options.Client().ApplyURI(dsn).SetServerAPIOptions(serverAPI)

// 	// 	client, err := mongo.Connect(ctx, opts)
// 	// 	if err != nil {
// 	// 		panic(err)
// 	// 	}

// 	// 	dbClient := client.Database(cfg.MongoDB.Name, options.Database().SetBSONOptions(&options.BSONOptions{
// 	// 		UseJSONStructTags:       true,
// 	// 		ErrorOnInlineDuplicates: false,
// 	// 		NilMapAsEmpty:           true,
// 	// 		NilSliceAsEmpty:         true,
// 	// 		NilByteSliceAsEmpty:     true,
// 	// 	}))

// 	// 	ctx = context.WithValue(ctx, dbClientKey, dbClient)
// 	// }

// 	// bskyClient, err := atproto.NewBlueskyClient(ctx, atproto.BlueskyClientConfig{
// 	// 	PDSURL:   cfg.Bluesky.PDSURL,
// 	// 	Username: cfg.Bluesky.Username,
// 	// 	Password: cfg.Bluesky.AppPassword,
// 	// })
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	// ctx = context.WithValue(ctx, bskyClientKey, bskyClient)

// 	// ctx = context.WithValue(ctx, loggerKey, logging.GetLogger(os.Stdout, cfg.LogLevel))

// 	// if cfg.HTTPServer.Enabled {
// 	// 	ctx = feedconfig.WithConfig(ctx, feedconfig.FeedGeneratorConfig{
// 	// 		ServiceDID: syntax.ATURI(cfg.Bluesky.FeedServiceDID),
// 	// 	})
// 	// }

// 	ctx, cancel := context.WithCancel(ctx)
// 	return ctx, cancel, nil
// }

// func AppConfig(ctx context.Context) *config.AppConfig {
// 	if ctx == nil {
// 		return nil
// 	}

// 	appCfg, ok := ctx.Value(appConfig).(*config.AppConfig)
// 	if !ok {
// 		return nil
// 	}

// 	return appCfg
// }

// func DBClient(ctx context.Context) *mongo.Database {
// 	if ctx == nil {
// 		return nil
// 	}

// 	dbClient, ok := ctx.Value(dbClientKey).(*mongo.Database)
// 	if !ok {
// 		return nil
// 	}

// 	return dbClient
// }

// func BlueskyClient(ctx context.Context) *atproto.BlueskyClient {
// 	if ctx == nil {
// 		return nil
// 	}

// 	bskyClient, ok := ctx.Value(bskyClientKey).(*atproto.BlueskyClient)
// 	if !ok {
// 		return nil
// 	}

// 	return bskyClient
// }

// func Logger(ctx context.Context) *slog.Logger {
// 	if ctx == nil {
// 		return nil
// 	}

// 	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
// 	if !ok {
// 		return nil
// 	}

// 	return logger
// }
