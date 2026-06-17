package runtime

import (
	"context"
	"net/http"

	"github.com/emersonary/appkit/accounts"
	appkitconfig "github.com/emersonary/appkit/config"
	"github.com/emersonary/appkit/currency"
	"google.golang.org/grpc"
)

// AccountsWireFunc supplies accounts.Options after infra is ready (e.g. product mailer).
type AccountsWireFunc[T appkitconfig.AppConfig] func(ctx context.Context, app *Application[T]) (accounts.Options, error)

// CurrencyWireFunc supplies currency.WireOptions after infra is ready.
type CurrencyWireFunc[T appkitconfig.AppConfig] func(ctx context.Context, app *Application[T], workerCtx context.Context) (currency.WireOptions, error)

// WireFunc registers domain services after infra (pool, NATS, health) is ready.
type WireFunc[T appkitconfig.AppConfig] func(ctx context.Context, app *Application[T]) error

// TransportWire is implemented by the product composition root (SaharApplication, etc.).
// Runtime calls it during createConnectionHandlers after WireServices completes.
type TransportWire[T appkitconfig.AppConfig] interface {
	NewGRPCServer() *grpc.Server
	WireRoutes(ctx context.Context, app *Application[T], mux *http.ServeMux, grpcSrv *grpc.Server) error
	WrapHTTP(http.Handler) http.Handler
}

// WireWebSocketFunc registers WebSocket routes on the shared HTTP mux.
type WireWebSocketFunc[T appkitconfig.AppConfig] func(ctx context.Context, app *Application[T], mux *http.ServeMux) error

// WireWebSocketServerFunc builds a handler for a dedicated WebSocket listener (server.websocket).
type WireWebSocketServerFunc[T appkitconfig.AppConfig] func(ctx context.Context, app *Application[T]) (http.Handler, error)

// WireShutdownFunc runs product cleanup after servers stop and before infra closes.
type WireShutdownFunc[T appkitconfig.AppConfig] func(ctx context.Context, app *Application[T]) error

// Options holds product wiring hooks; product config type is T.
type Options[T appkitconfig.AppConfig] struct {
	AccountsWire        AccountsWireFunc[T]
	CurrencyWire        CurrencyWireFunc[T]
	WireServices        WireFunc[T]
	Transport           TransportWire[T]
	WireWebSocket       WireWebSocketFunc[T]
	WireWebSocketServer WireWebSocketServerFunc[T]
	WireShutdown        WireShutdownFunc[T]
}
