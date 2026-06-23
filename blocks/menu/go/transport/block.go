package transport

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/grpc"

	menuv1 "github.com/emersonary/appkit/menu/gen/menu/v1"
	"github.com/emersonary/appkit/menu/gen/menu/v1/menuv1connect"
	"github.com/emersonary/appkit/menu"
)

// Block wires menu transport around a Service.
type Block struct {
	Service    *menu.Service
	menuServer *MenuServer
}

// Mount configures where menu transport is registered.
type Mount struct {
	HTTPMux          *http.ServeMux
	GRPCServer       *grpc.Server
	ConnectOptions   []connect.HandlerOption
	ResolveAccountID func(ctx context.Context, accessToken string) (string, error)
}

// New builds the menu transport block and registers endpoints when mount is non-nil.
func New(svc *menu.Service, mount *Mount) *Block {
	var accountID func(context.Context) (string, error)
	if mount != nil && mount.ResolveAccountID != nil {
		resolve := mount.ResolveAccountID
		accountID = func(ctx context.Context) (string, error) {
			if id, ok := AccountIDFromContext(ctx); ok {
				return id, nil
			}
			return AccountIDFromBearer(ctx, resolve)
		}
	} else {
		accountID = func(ctx context.Context) (string, error) {
			if id, ok := AccountIDFromContext(ctx); ok {
				return id, nil
			}
			return "", MapGRPCError(menu.ErrUnauthenticated)
		}
	}

	b := &Block{
		Service:    svc,
		menuServer: NewMenuServer(svc, accountID),
	}
	if mount != nil {
		b.mount(mount)
	}
	return b
}

func (b *Block) ConnectHandler(opts ...connect.HandlerOption) (mountPath string, handler http.Handler) {
	return menuv1connect.NewMenuServiceHandler(newConnectMenuService(b.menuServer), opts...)
}

func (b *Block) mount(m *Mount) {
	if m.HTTPMux != nil {
		path, handler := menuv1connect.NewMenuServiceHandler(
			newConnectMenuService(b.menuServer),
			m.ConnectOptions...,
		)
		m.HTTPMux.Handle(path, handler)
	}
	if m.GRPCServer != nil {
		menuv1.RegisterMenuServiceServer(m.GRPCServer, b.menuServer)
	}
}
