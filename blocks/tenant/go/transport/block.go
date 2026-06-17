package transport

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/grpc"

	tenantv1 "github.com/emersonary/appkit/tenants/gen/tenant/v1"
	"github.com/emersonary/appkit/tenants/gen/tenant/v1/tenantv1connect"
	"github.com/emersonary/appkit/tenants"
)

// Block wires tenant transport (Connect, gRPC) around a Service.
type Block struct {
	Service      *tenants.Service
	tenantServer *TenantServer
}

// Mount configures where tenant transport is registered.
type Mount struct {
	HTTPMux            *http.ServeMux
	GRPCServer         *grpc.Server
	ConnectOptions     []connect.HandlerOption
	ResolveAccountID   AccountIDResolver
}

// New builds the tenant transport block and registers endpoints when mount is non-nil.
func New(svc *tenants.Service, mount *Mount) *Block {
	b := &Block{
		Service:      svc,
		tenantServer: NewTenantServer(svc),
	}
	if mount != nil {
		b.mount(mount)
	}
	return b
}

func (b *Block) mount(m *Mount) {
	auth := connect.WithInterceptors(requireAccountSession(m.ResolveAccountID))
	opts := append([]connect.HandlerOption{auth}, m.ConnectOptions...)

	if m.HTTPMux != nil {
		path, handler := tenantv1connect.NewTenantServiceHandler(newConnectTenantService(b.tenantServer), opts...)
		m.HTTPMux.Handle(path, handler)
	}
	if m.GRPCServer != nil {
		tenantv1.RegisterTenantServiceServer(m.GRPCServer, b.tenantServer)
	}
}

func requireAccountSession(resolver AccountIDResolver) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if resolver == nil {
				return nil, ToConnectError(tenants.ErrUnauthenticated)
			}
			ctx = ContextWithAuthorizationHeader(ctx, req.Header().Get("Authorization"))
			nextCtx, err := Authenticator(resolver)(ctx)
			if err != nil {
				return nil, ToConnectError(err)
			}
			return next(nextCtx, req)
		}
	}
}
