package runtime

import (
	"context"
	"net"
	"net/http"

	"github.com/emersonary/appkit/accounts"
	"github.com/emersonary/appkit/ai"
	appkitconfig "github.com/emersonary/appkit/config"
	"github.com/emersonary/appkit/currency"
	"github.com/emersonary/appkit/dbhist"
	"github.com/emersonary/appkit/email"
	"github.com/emersonary/appkit/health"
	"github.com/emersonary/appkit/language"
	"github.com/emersonary/appkit/menu"
	"github.com/emersonary/appkit/permissions"
	"github.com/emersonary/appkit/tenants"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Application is the shared runtime shell: infra, health, and HTTP/gRPC/WebSocket servers.
// T is the product config type (e.g. sahar config.Config) embedding appkitconfig.BaseConfig.
type Application[T appkitconfig.AppConfig] struct {
	Config T

	Logger *zap.Logger
	Pool   *pgxpool.Pool
	NATS   *nats.Conn
	Redis  *redis.Client

	httpServer   *http.Server
	grpcServer   *grpc.Server
	grpcListener net.Listener
	wsServer     *http.Server

	workerCancels []context.CancelFunc
	wireShutdown  WireShutdownFunc[T]

	Health      *health.Service
	Accounts    *accounts.Service
	Tenants     *tenants.Service
	Currency    *currency.Service
	Language    *language.Service
	AI          *ai.Service
	Permissions *permissions.Service
	Menu        *menu.Service
	Email       *email.Service
	DBHist      *dbhist.Service
}

// Base returns the embedded appkit infra config from the product config.
func (a *Application[T]) Base() *appkitconfig.BaseConfig {
	return a.Config.Infra()
}

// New builds logger, database, messaging, health, product services, and transport.
func New[T appkitconfig.AppConfig](ctx context.Context, cfg T, opts Options[T]) (*Application[T], error) {
	app := &Application[T]{
		Config:       cfg,
		wireShutdown: opts.WireShutdown,
	}

	base := cfg.Infra()

	if err := app.createLogger(base.Log); err != nil {
		return nil, err
	}

	if err := app.createDatabaseConnection(ctx); err != nil {
		app.Shutdown(context.Background())
		return nil, err
	}

	if base.NATS.URL != "" {
		if err := app.createNATSConnection(base.NATS.URL); err != nil {
			app.Shutdown(context.Background())
			return nil, err
		}
	}

	if base.Redis.Enabled {
		if err := app.createRedisConnection(ctx); err != nil {
			app.Shutdown(context.Background())
			return nil, err
		}
	}

	app.createHealthService()
	app.wireEmail()

	if err := app.wireBlocks(ctx, opts); err != nil {
		app.Shutdown(context.Background())
		return nil, err
	}

	if opts.WireServices != nil {
		if err := opts.WireServices(ctx, app); err != nil {
			app.Shutdown(context.Background())
			return nil, err
		}
	}

	if err := app.createConnectionHandlers(ctx, opts); err != nil {
		app.Shutdown(context.Background())
		return nil, err
	}

	return app, nil
}

// Start serves enabled HTTP, gRPC, and WebSocket listeners in background goroutines.
func (a *Application[T]) Start() {
	if a.grpcServer != nil && a.grpcListener != nil {
		addr := a.Base().Server.GRPC.Addr
		go func() {
			a.Logger.Info("grpc listening", zap.String("addr", addr))
			if err := a.grpcServer.Serve(a.grpcListener); err != nil {
				a.LogError("grpc serve", err)
			}
		}()
	}

	if a.httpServer != nil {
		addr := a.Base().Server.HTTP.Addr
		go func() {
			a.Logger.Info("http listening", zap.String("addr", addr))
			if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				a.LogError("http serve", err)
			}
		}()
	}

	if a.wsServer != nil {
		go func() {
			a.Logger.Info("websocket listening", zap.String("addr", a.wsServer.Addr))
			if err := a.wsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				a.LogError("websocket serve", err)
			}
		}()
	}
}

// Shutdown gracefully stops servers and closes infrastructure connections.
func (a *Application[T]) Shutdown(ctx context.Context) error {
	var shutdownErr error

	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			shutdownErr = err
		}
		a.httpServer = nil
	}
	if a.wsServer != nil {
		if err := a.wsServer.Shutdown(ctx); err != nil && shutdownErr == nil {
			shutdownErr = err
		}
		a.wsServer = nil
	}
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
		a.grpcServer = nil
	}
	if a.grpcListener != nil {
		_ = a.grpcListener.Close()
		a.grpcListener = nil
	}
	if a.wireShutdown != nil {
		if err := a.wireShutdown(ctx, a); err != nil && shutdownErr == nil {
			shutdownErr = err
		}
	}
	a.stopWorkers()
	if a.NATS != nil {
		a.NATS.Close()
		a.NATS = nil
	}
	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil && shutdownErr == nil {
			shutdownErr = err
		}
		a.Redis = nil
	}
	if a.Pool != nil {
		a.Pool.Close()
		a.Pool = nil
	}

	return shutdownErr
}
