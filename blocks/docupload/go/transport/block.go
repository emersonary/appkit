package transport

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/grpc"

	"github.com/emersonary/appkit/docupload"
	docuploadv1 "github.com/emersonary/appkit/docupload/gen/docupload/v1"
	"github.com/emersonary/appkit/docupload/gen/docupload/v1/docuploadv1connect"
)

// Block wires doc upload transport around a Service.
type Block struct {
	Service *docupload.Service
	server  *Server
}

// Mount configures where doc upload transport is registered.
type Mount struct {
	HTTPMux          *http.ServeMux
	GRPCServer       *grpc.Server
	ConnectOptions   []connect.HandlerOption
	ResolveAccountID AccountIDResolver
}

// New builds the transport block and registers endpoints when mount is non-nil.
func New(svc *docupload.Service, mount *Mount) *Block {
	b := &Block{Service: svc, server: NewServer(svc)}
	if mount != nil {
		b.mount(mount)
	}
	return b
}

func (b *Block) mount(m *Mount) {
	auth := connect.WithInterceptors(requireAccountSession(m.ResolveAccountID))
	opts := append([]connect.HandlerOption{auth}, m.ConnectOptions...)
	if m.HTTPMux != nil {
		path, handler := docuploadv1connect.NewDocUploadServiceHandler(newConnectService(b.server), opts...)
		m.HTTPMux.Handle(path, handler)
	}
	if m.GRPCServer != nil {
		docuploadv1.RegisterDocUploadServiceServer(m.GRPCServer, b.server)
	}
}

func requireAccountSession(resolver AccountIDResolver) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if resolver == nil {
				return nil, ToConnectError(docupload.ErrUnauthenticated)
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
