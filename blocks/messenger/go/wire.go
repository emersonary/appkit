package messenger

import (
	"context"
	"database/sql"

	"github.com/emersonary/appkit/messenger/ws"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type WireOptions struct {
	Logger *zap.Logger
	NATS   *nats.Conn
	Hub    *ws.Hub
}

type Block struct {
	Service    *Service
	Hub        *ws.Hub
	Subscriber *Subscriber
}

func ApplySchemaFromAppConfig(ctx context.Context, db *sql.DB, app AppConfig) error {
	if !app.Enabled {
		return nil
	}
	cfg, err := ResolveBlockConfig(app)
	if err != nil {
		return err
	}
	return ApplySchema(ctx, db, cfg)
}

func Wire(ctx context.Context, db *sql.DB, app AppConfig, opts WireOptions) (*Block, error) {
	if !app.Enabled {
		return nil, nil
	}

	cfg, err := ResolveBlockConfig(app)
	if err != nil {
		return nil, err
	}

	if err := ApplySchema(ctx, db, cfg); err != nil {
		return nil, err
	}

	store := NewStore(db, cfg)
	var publisher *Publisher
	if opts.NATS != nil {
		publisher = NewPublisher(opts.NATS, opts.Logger)
	}

	svc := NewService(store, cfg, publisher, opts.Logger)

	hub := opts.Hub
	if hub == nil {
		hub = ws.NewHub()
	}

	var sub *Subscriber
	if opts.NATS != nil {
		sub = NewSubscriber(opts.NATS, hub, opts.Logger)
		if err := sub.Start(ctx); err != nil {
			return nil, err
		}
	}

	if opts.Logger != nil {
		opts.Logger.Named("messenger").Info("messenger block enabled")
	}

	return &Block{Service: svc, Hub: hub, Subscriber: sub}, nil
}
