package transport

import (
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/grpc"

	accountv1 "github.com/emersonary/appkit/accounts/gen/account/v1"
	"github.com/emersonary/appkit/accounts/gen/account/v1/accountv1connect"
	accounthttp "github.com/emersonary/appkit/accounts/http"
	"github.com/emersonary/appkit/accounts"
)

// Block wires account transport (Connect, gRPC, default REST) around a Service.
type Block struct {
	Service       *accounts.Service
	accountServer *AccountServer
}

// HTTPRoute registers an optional extra REST or WebSocket handler owned by the account block.
type HTTPRoute struct {
	Pattern string
	Handler http.Handler
}

// Mount configures where account transport is registered. Fields are optional; pass nil to
// NewBlock when only the service wrapper is needed (e.g. unit tests).
type Mount struct {
	HTTPMux        *http.ServeMux
	GRPCServer     *grpc.Server
	ConnectOptions []connect.HandlerOption
	ExtraRoutes    []HTTPRoute
}

// NewBlock builds the account transport block and, when mount is non-nil, registers all
// account endpoints on the given HTTP mux and/or gRPC server.
func NewBlock(svc *accounts.Service, mount *Mount) *Block {
	b := &Block{
		Service:       svc,
		accountServer: NewAccountServer(svc),
	}
	if mount != nil {
		b.mount(mount)
	}
	return b
}

// AccountServer returns the gRPC/Connect implementation for advanced wiring.
func (b *Block) AccountServer() *AccountServer {
	return b.accountServer
}

// ConnectHandler exposes the account Connect service as an http.Handler.
func (b *Block) ConnectHandler(opts ...connect.HandlerOption) (mountPath string, handler http.Handler) {
	return accountv1connect.NewAccountServiceHandler(newConnectAccountService(b.accountServer), opts...)
}

func (b *Block) mount(m *Mount) {
	if m.HTTPMux != nil {
		for _, route := range m.ExtraRoutes {
			m.HTTPMux.Handle(route.Pattern, route.Handler)
		}
		accounthttp.New(b.Service).RegisterRoutes(m.HTTPMux)
		path, handler := accountv1connect.NewAccountServiceHandler(
			newConnectAccountService(b.accountServer),
			m.ConnectOptions...,
		)
		m.HTTPMux.Handle(path, handler)
	}
	if m.GRPCServer != nil {
		accountv1.RegisterAccountServiceServer(m.GRPCServer, b.accountServer)
	}
}
