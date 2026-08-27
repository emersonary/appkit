package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/emersonary/appkit/db/querylog"
	"github.com/emersonary/appkit/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ApplyConfig is the product migration registry passed to appkit migrate.
type ApplyConfig = migrate.ApplyConfig

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	return NewPoolWithLogger(ctx, dsn, nil)
}

func NewPoolWithLogger(ctx context.Context, dsn string, logger *zap.Logger) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	poolCfg.MaxConns = 10
	poolCfg.MinConns = 1
	poolCfg.MaxConnLifetime = time.Hour
	if logger != nil {
		poolCfg.ConnConfig.Tracer = &querylog.Tracer{Logger: logger.Named("db")}
	}

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
	pool, err := NewPoolWithLogger(ctx, a.Base().Database.DSN(), a.Logger)
	if err != nil {
		return err
	}

	a.Pool = pool
	a.Logger.Info("postgres connected")

	if path := a.Base().Migrations.Path; path != "" {
		if err := migrate.RunGooseUp(ctx, pool, path); err != nil {
			return err
		}
	}

	cfg := migrate.ApplyConfig{}
	if a.migrationsWire != nil {
		cfg, err = a.migrationsWire(ctx, a.Config)
		if err != nil {
			return err
		}
	}

	return migrate.RunMigrations(ctx, pool, cfg)
}
