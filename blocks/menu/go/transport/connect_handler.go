package transport

import (
	"context"

	"connectrpc.com/connect"

	menuv1 "github.com/emersonary/appkit/menu/gen/menu/v1"
)

type connectMenuService struct {
	inner *MenuServer
}

func newConnectMenuService(inner *MenuServer) *connectMenuService {
	return &connectMenuService{inner: inner}
}

func (h *connectMenuService) GetMenu(
	ctx context.Context,
	req *connect.Request[menuv1.GetMenuRequest],
) (*connect.Response[menuv1.GetMenuResponse], error) {
	ctx = ContextWithAuthorizationHeader(ctx, req.Header().Get("Authorization"))
	resp, err := h.inner.GetMenu(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}
