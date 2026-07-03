package weather

import (
	"context"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type WireOptions struct {
	Redis      *redis.Client
	Logger     *zap.Logger
	WorkerCtx  context.Context
	HTTPClient *http.Client
}

func Wire(ctx context.Context, cfg AppConfig, opts WireOptions) (*Service, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if opts.Redis == nil {
		return nil, ErrRedisRequired
	}

	resolved, err := ResolveBlockConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("resolve weather config: %w", err)
	}

	store := NewRedisStore(opts.Redis, resolved.KeyPrefix, resolved.CacheTTL)
	client := NewOpenMeteoClient(resolved.OpenMeteo, opts.HTTPClient)
	svc, err := NewService(resolved, client, store, Options{Logger: opts.Logger})
	if err != nil {
		return nil, err
	}

	if opts.WorkerCtx != nil {
		go svc.RunCollector(opts.WorkerCtx, resolved.RefreshInterval)
	}
	_ = ctx
	return svc, nil
}
