package appcontext

import (
	"context"
	"errors"
	"sync"

	"github.com/jghiloni/watchedsky-social/backend/config"
)

type ContextLoader func(context.Context, config.AppConfig) (context.Context, error)

type clientRegistry struct {
	loaders []ContextLoader
	mu      *sync.RWMutex
	loaded  bool
}

func (c *clientRegistry) RegisterClient(a ContextLoader) {
	if c.loaded {
		panic("clients already loaded")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.loaders = append(c.loaders, a)
}

func (c *clientRegistry) LoadClients(ctx context.Context) (context.Context, error) {
	cfg := config.GetConfig(ctx)

	clientErrors := []error{}
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, loader := range c.loaders {
		newCtx, err := loader(ctx, cfg)
		if err != nil {
			clientErrors = append(clientErrors, err)
			continue
		}
		ctx = newCtx
	}

	if len(clientErrors) == 0 {
		c.loaded = true
		return ctx, nil
	}

	return nil, errors.Join(clientErrors...)
}

var Registry *clientRegistry = &clientRegistry{
	loaders: []ContextLoader{},
	mu:      &sync.RWMutex{},
	loaded:  false,
}
