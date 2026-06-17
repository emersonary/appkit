package runtime

import (
	"context"
	"net"
	"net/http"
	"time"

	appkitconfig "github.com/emersonary/appkit/config"
	"google.golang.org/grpc"
)

func (a *Application[T]) createConnectionHandlers(ctx context.Context, opts Options[T]) error {
	base := a.Base()
	httpEnabled := base.Server.HTTP.Active()
	grpcEnabled := base.Server.GRPC.Active()
	wsEnabled := base.Server.WebSocket.Active()

	if httpEnabled || grpcEnabled {
		var httpMux *http.ServeMux
		var grpcSrv *grpc.Server

		if grpcEnabled {
			if opts.Transport != nil {
				grpcSrv = opts.Transport.NewGRPCServer()
			} else {
				grpcSrv = grpc.NewServer()
			}
		}

		if httpEnabled {
			httpMux = http.NewServeMux()
			a.HealthHandler.RegisterRoutes(httpMux)
			appkitconfig.NewHandler(func() any { return a.Config }).RegisterRoutes(httpMux, "GET /config")
		}

		if opts.Transport != nil {
			if err := opts.Transport.WireRoutes(ctx, a, httpMux, grpcSrv); err != nil {
				return err
			}
		}

		if httpEnabled && opts.WireWebSocket != nil {
			if err := opts.WireWebSocket(ctx, a, httpMux); err != nil {
				return err
			}
		}

		if grpcEnabled {
			grpcListener, err := net.Listen("tcp", base.Server.GRPC.Addr)
			if err != nil {
				return err
			}
			a.grpcServer = grpcSrv
			a.grpcListener = grpcListener
		}

		if httpEnabled {
			httpHandler := http.Handler(httpMux)
			if opts.Transport != nil {
				httpHandler = opts.Transport.WrapHTTP(httpHandler)
			}
			a.httpServer = &http.Server{
				Addr:              base.Server.HTTP.Addr,
				Handler:           httpHandler,
				ReadHeaderTimeout: 10 * time.Second,
			}
		}
	}

	if wsEnabled && opts.WireWebSocketServer != nil {
		wsHandler, err := opts.WireWebSocketServer(ctx, a)
		if err != nil {
			return err
		}
		a.wsServer = &http.Server{
			Addr:              base.Server.WebSocket.Addr,
			Handler:           wsHandler,
			ReadHeaderTimeout: 10 * time.Second,
		}
	}

	return nil
}
