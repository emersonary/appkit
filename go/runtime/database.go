package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/emersonary/appkit/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ApplyConfig is the product migration registry passed to appkit migrate.
type ApplyConfig = migrate.ApplyConfig

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	poolCfg.MaxConns = 10
	poolCfg.MinConns = 1
	poolCfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

func (a *Application[T]) createDatabaseConnection(ctx context.Context) error {
	pool, err := NewPool(ctx, a.Base().Database.DSN())
	if err != nil {
		return err
	}

	a.Pool = pool
	a.Logger.Info("postgres connected")

	cfg := migrate.ApplyConfig{}
	if a.migrationsWire != nil {
		cfg, err = a.migrationsWire(ctx, a.Config)
		if err != nil {
			return err
		}
	}

	return migrate.RunMigrations(ctx, pool, cfg)
}
