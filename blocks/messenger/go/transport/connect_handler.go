package transport

import (
	"context"

	"connectrpc.com/connect"

	messengerv1 "github.com/emersonary/appkit/messenger/gen/messenger/v1"
)

type connectMessengerService struct {
	inner *MessengerServer
}

func newConnectMessengerService(inner *MessengerServer) *connectMessengerService {
	return &connectMessengerService{inner: inner}
}

func (h *connectMessengerService) ResolveMe(ctx context.Context, req *connect.Request[messengerv1.ResolveMeRequest]) (*connect.Response[messengerv1.ResolveMeResponse], error) {
	resp, err := h.inner.ResolveMe(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectMessengerService) CreateAgentUser(ctx context.Context, req *connect.Request[messengerv1.CreateAgentUserRequest]) (*connect.Response[messengerv1.ResolveMeResponse], error) {
	resp, err := h.inner.CreateAgentUser(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectMessengerService) GetAgentUser(ctx context.Context, req *connect.Request[messengerv1.GetAgentUserRequest]) (*connect.Response[messengerv1.ResolveMeResponse], error) {
	resp, err := h.inner.GetAgentUser(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectMessengerService) RegisterDevice(ctx context.Context, req *connect.Request[messengerv1.RegisterDeviceRequest]) (*connect.Response[messengerv1.RegisterDeviceResponse], error) {
	resp, err := h.inner.RegisterDevice(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectMessengerService) GetOrCreateDirect(ctx context.Context, req *connect.Request[messengerv1.GetOrCreateDirectRequest]) (*connect.Response[messengerv1.ChannelResponse], error) {
	resp, err := h.inner.GetOrCreateDirect(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectMessengerService) CreateGroup(ctx context.Context, req *connect.Request[messengerv1.CreateGroupRequest]) (*connect.Response[messengerv1.ChannelResponse], error) {
	resp, err := h.inner.CreateGroup(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectMessengerService) ListChannels(ctx context.Context, req *connect.Request[messengerv1.ListChannelsRequest]) (*connect.Response[messengerv1.ListChannelsResponse], error) {
	resp, err := h.inner.ListChannels(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectMessengerService) ListMessages(ctx context.Context, req *connect.Request[messengerv1.ListMessagesRequest]) (*connect.Response[messengerv1.ListMessagesResponse], error) {
	resp, err := h.inner.ListMessages(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectMessengerService) SendMessage(ctx context.Context, req *connect.Request[messengerv1.SendMessageRequest]) (*connect.Response[messengerv1.Message], error) {
	resp, err := h.inner.SendMessage(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectMessengerService) SendSystemMessage(ctx context.Context, req *connect.Request[messengerv1.SendSystemMessageRequest]) (*connect.Response[messengerv1.Message], error) {
	resp, err := h.inner.SendSystemMessage(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectMessengerService) MarkRead(ctx context.Context, req *connect.Request[messengerv1.MarkReadRequest]) (*connect.Response[messengerv1.MarkReadResponse], error) {
	resp, err := h.inner.MarkRead(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectMessengerService) SyncMessages(ctx context.Context, req *connect.Request[messengerv1.SyncMessagesRequest]) (*connect.Response[messengerv1.SyncMessagesResponse], error) {
	resp, err := h.inner.SyncMessages(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectMessengerService) HandoffConversation(ctx context.Context, req *connect.Request[messengerv1.HandoffConversationRequest]) (*connect.Response[messengerv1.HandoffConversationResponse], error) {
	resp, err := h.inner.HandoffConversation(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}
