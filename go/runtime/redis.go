package runtime

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func (a *Application[T]) createRedisConnection(ctx context.Context) error {
	cfg := a.Base().Redis
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return err
	}

	a.Redis = client
	a.Logger.Info("redis connected", zap.String("addr", cfg.Addr), zap.Int("db", cfg.DB))
	return nil
}
