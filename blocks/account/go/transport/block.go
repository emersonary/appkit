package transport

import (
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/grpc"

	accountv1 "github.com/emersonary/appkit/accounts/gen/account/v1"
	"github.com/emersonary/appkit/accounts/gen/account/v1/accountv1connect"
	accounthttp "github.com/emersonary/appkit/accounts/http"
	"github.com/emersonary/appkit/accounts"
	"github.com/emersonary/appkit/accounts/oauth"
)

// Block wires account transport (Connect, gRPC, default REST) around a Service.
type Block struct {
	Service       *accounts.Service
	accountServer *AccountServer
}

func NewBlock(svc *accounts.Service) *Block {
	return &Block{
		Service:       svc,
		accountServer: NewAccountServer(svc),
	}
}

// HTTPRoute registers an extra REST or WebSocket handler on the shared mux.
type HTTPRoute struct {
	Pattern string
	Handler http.Handler
}

// HTTPMount configures HTTP registration for the block.
type HTTPMount struct {
	OAuthProvider  oauth.Provider
	ConnectOptions []connect.HandlerOption
	ExtraRoutes    []HTTPRoute
	// RegisterExtra mounts app-specific REST or WebSocket routes before block defaults.
	RegisterExtra func(mux *http.ServeMux)
}

// AccountServer returns the gRPC/Connect implementation for advanced wiring.
func (b *Block) AccountServer() *AccountServer {
	return b.accountServer
}

// ConnectHandler exposes the account Connect service as an http.Handler.
func (b *Block) ConnectHandler(opts ...connect.HandlerOption) (mountPath string, handler http.Handler) {
	return accountv1connect.NewAccountServiceHandler(newConnectAccountService(b.accountServer), opts...)
}

// MountHTTP registers default account REST routes, Connect, and any extra routes.
func (b *Block) MountHTTP(mux *http.ServeMux, mount HTTPMount) {
	if mount.RegisterExtra != nil {
		mount.RegisterExtra(mux)
	}
	for _, route := range mount.ExtraRoutes {
		mux.Handle(route.Pattern, route.Handler)
	}
	accounthttp.New(b.Service, mount.OAuthProvider).RegisterRoutes(mux)
	path, handler := accountv1connect.NewAccountServiceHandler(
		newConnectAccountService(b.accountServer),
		mount.ConnectOptions...,
	)
	mux.Handle(path, handler)
}

// RegisterGRPC registers AccountService on a gRPC server.
func (b *Block) RegisterGRPC(grpcServer *grpc.Server) {
	accountv1.RegisterAccountServiceServer(grpcServer, b.accountServer)
}
