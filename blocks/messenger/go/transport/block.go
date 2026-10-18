package transport

import (
	"context"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"google.golang.org/grpc"

	messengerv1 "github.com/emersonary/appkit/messenger/gen/messenger/v1"
	"github.com/emersonary/appkit/messenger/gen/messenger/v1/messengerv1connect"
	"github.com/emersonary/appkit/messenger"
)

type Block struct {
	Service         *messenger.Service
	messengerServer *MessengerServer
	wsHandler       *WSHandler
	agentAPIKey     string
}

type Mount struct {
	HTTPMux          *http.ServeMux
	GRPCServer       *grpc.Server
	ConnectOptions   []connect.HandlerOption
	ResolveAccountID AccountIDResolver
	AgentAPIKey      string
}

func New(block *messenger.Block, mount *Mount) *Block {
	if block == nil || block.Service == nil {
		return nil
	}

	agentKey := ""
	var resolve AccountIDResolver
	if mount != nil {
		agentKey = mount.AgentAPIKey
		resolve = mount.ResolveAccountID
	}

	b := &Block{
		Service:         block.Service,
		messengerServer: NewMessengerServer(block.Service),
		agentAPIKey:     agentKey,
		wsHandler:       NewWSHandler(block, resolve, agentKey),
	}
	if mount != nil {
		b.mount(mount)
	}
	return b
}

func (b *Block) mount(m *Mount) {
	opts := append([]connect.HandlerOption{b.authInterceptor(m.ResolveAccountID, m.AgentAPIKey)}, m.ConnectOptions...)

	if m.HTTPMux != nil {
		path, handler := messengerv1connect.NewMessengerServiceHandler(newConnectMessengerService(b.messengerServer), opts...)
		m.HTTPMux.Handle(path, handler)
		b.wsHandler.Register(m.HTTPMux)
	}
	if m.GRPCServer != nil {
		messengerv1.RegisterMessengerServiceServer(m.GRPCServer, b.messengerServer)
	}
}

func (b *Block) authInterceptor(resolve AccountIDResolver, agentAPIKey string) connect.HandlerOption {
	return connect.WithInterceptors(connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			ctx = ContextWithAuthorizationHeader(ctx, req.Header().Get("Authorization"))

			if agentKey := strings.TrimSpace(req.Header().Get("X-Messenger-Agent-Key")); agentKey != "" && agentAPIKey != "" && agentKey == agentAPIKey {
				ctx = WithAgentAuth(ctx)
				if agentUserID := strings.TrimSpace(req.Header().Get("X-Messenger-Agent-User-Id")); agentUserID != "" {
					ctx = WithMessengerUserID(ctx, agentUserID)
				}
				return next(ctx, req)
			}

			if resolve == nil {
				return nil, ToConnectError(messenger.ErrUnauthenticated)
			}
			nextCtx, err := Authenticator(resolve)(ctx)
			if err != nil {
				return nil, ToConnectError(err)
			}
			return next(nextCtx, req)
		}
	}))
}
