package transport

import (
	"context"

	messengerv1 "github.com/emersonary/appkit/messenger/gen/messenger/v1"
	"github.com/emersonary/appkit/messenger"
)

type MessengerServer struct {
	messengerv1.UnimplementedMessengerServiceServer
	svc *messenger.Service
}

func NewMessengerServer(svc *messenger.Service) *MessengerServer {
	return &MessengerServer{svc: svc}
}

func (s *MessengerServer) ResolveMe(ctx context.Context, req *messengerv1.ResolveMeRequest) (*messengerv1.ResolveMeResponse, error) {
	accountID, ok := AccountIDFromContext(ctx)
	if !ok {
		return nil, MapGRPCError(messenger.ErrUnauthenticated)
	}
	user, err := s.svc.ResolveHumanUser(ctx, req.GetTenantId(), accountID, req.GetDisplayName())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return &messengerv1.ResolveMeResponse{User: toProtoUser(user)}, nil
}

func (s *MessengerServer) CreateAgentUser(ctx context.Context, req *messengerv1.CreateAgentUserRequest) (*messengerv1.ResolveMeResponse, error) {
	if !AgentAuthFromContext(ctx) {
		return nil, MapGRPCError(messenger.ErrForbidden)
	}
	user, err := s.svc.CreateAgentUser(ctx, req.GetTenantId(), req.GetDisplayName(), req.GetExternalRef(), req.GetAvatarUrl())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return &messengerv1.ResolveMeResponse{User: toProtoUser(user)}, nil
}

func (s *MessengerServer) GetAgentUser(ctx context.Context, req *messengerv1.GetAgentUserRequest) (*messengerv1.ResolveMeResponse, error) {
	if !AgentAuthFromContext(ctx) {
		return nil, MapGRPCError(messenger.ErrForbidden)
	}
	user, err := s.svc.GetAgentUser(ctx, req.GetTenantId(), req.GetExternalRef())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return &messengerv1.ResolveMeResponse{User: toProtoUser(user)}, nil
}

func (s *MessengerServer) RegisterDevice(ctx context.Context, req *messengerv1.RegisterDeviceRequest) (*messengerv1.RegisterDeviceResponse, error) {
	userID, err := s.resolveActorUser(ctx, req.GetTenantId())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	dev, err := s.svc.RegisterDevice(ctx, userID, req.GetDeviceKey(), req.GetPlatform(), req.GetPushToken())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return &messengerv1.RegisterDeviceResponse{DeviceId: dev.ID}, nil
}

func (s *MessengerServer) GetOrCreateDirect(ctx context.Context, req *messengerv1.GetOrCreateDirectRequest) (*messengerv1.ChannelResponse, error) {
	userID, err := s.resolveActorUser(ctx, req.GetTenantId())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	ch, err := s.svc.GetOrCreateDirect(ctx, req.GetTenantId(), userID, req.GetOtherUserId())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return &messengerv1.ChannelResponse{Channel: toProtoChannel(ch)}, nil
}

func (s *MessengerServer) CreateGroup(ctx context.Context, req *messengerv1.CreateGroupRequest) (*messengerv1.ChannelResponse, error) {
	userID, err := s.resolveActorUser(ctx, req.GetTenantId())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	ch, err := s.svc.CreateGroup(ctx, req.GetTenantId(), userID, req.GetTitle(), req.GetMemberUserIds())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return &messengerv1.ChannelResponse{Channel: toProtoChannel(ch)}, nil
}

func (s *MessengerServer) ListChannels(ctx context.Context, req *messengerv1.ListChannelsRequest) (*messengerv1.ListChannelsResponse, error) {
	userID, err := s.resolveActorUser(ctx, req.GetTenantId())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	channels, err := s.svc.ListChannels(ctx, req.GetTenantId(), userID)
	if err != nil {
		return nil, MapGRPCError(err)
	}
	out := make([]*messengerv1.Channel, 0, len(channels))
	for _, ch := range channels {
		out = append(out, toProtoChannel(ch))
	}
	return &messengerv1.ListChannelsResponse{Channels: out}, nil
}

func (s *MessengerServer) ListMessages(ctx context.Context, req *messengerv1.ListMessagesRequest) (*messengerv1.ListMessagesResponse, error) {
	userID, err := s.resolveActorUser(ctx, req.GetTenantId())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	msgs, err := s.svc.ListMessages(ctx, req.GetTenantId(), userID, req.GetChannelId(), req.GetBeforeMessageId(), int(req.GetLimit()))
	if err != nil {
		return nil, MapGRPCError(err)
	}
	out := make([]*messengerv1.Message, 0, len(msgs))
	for _, msg := range msgs {
		out = append(out, toProtoMessage(msg))
	}
	return &messengerv1.ListMessagesResponse{Messages: out}, nil
}

func (s *MessengerServer) SendMessage(ctx context.Context, req *messengerv1.SendMessageRequest) (*messengerv1.Message, error) {
	userID, err := s.resolveActorUser(ctx, req.GetTenantId())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	msg, err := s.svc.SendUserMessage(ctx, req.GetTenantId(), userID, req.GetChannelId(), req.GetBody(), req.GetClientRef(), req.GetReplyToId(), req.GetDeviceId())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return toProtoMessage(msg), nil
}

func (s *MessengerServer) SendSystemMessage(ctx context.Context, req *messengerv1.SendSystemMessageRequest) (*messengerv1.Message, error) {
	if !AgentAuthFromContext(ctx) {
		return nil, MapGRPCError(messenger.ErrForbidden)
	}
	msg, err := s.svc.SendSystemMessage(ctx, req.GetTenantId(), req.GetChannelId(), req.GetEventType(), req.GetBody(), structToRaw(req.GetEventPayload()), nil)
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return toProtoMessage(msg), nil
}

func (s *MessengerServer) MarkRead(ctx context.Context, req *messengerv1.MarkReadRequest) (*messengerv1.MarkReadResponse, error) {
	userID, err := s.resolveActorUser(ctx, req.GetTenantId())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	if err := s.svc.MarkRead(ctx, req.GetTenantId(), userID, req.GetChannelId(), req.GetMessageId()); err != nil {
		return nil, MapGRPCError(err)
	}
	return &messengerv1.MarkReadResponse{}, nil
}

func (s *MessengerServer) SyncMessages(ctx context.Context, req *messengerv1.SyncMessagesRequest) (*messengerv1.SyncMessagesResponse, error) {
	userID, err := s.resolveActorUser(ctx, req.GetTenantId())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	results, err := s.svc.SyncMessages(ctx, req.GetTenantId(), userID, req.GetChannelCursors())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	out := make([]*messengerv1.ChannelSyncEntry, 0, len(results))
	for _, r := range results {
		entry := &messengerv1.ChannelSyncEntry{Channel: toProtoChannel(r.Channel)}
		for _, msg := range r.Messages {
			entry.Messages = append(entry.Messages, toProtoMessage(msg))
		}
		out = append(out, entry)
	}
	return &messengerv1.SyncMessagesResponse{Entries: out}, nil
}

func (s *MessengerServer) HandoffConversation(ctx context.Context, req *messengerv1.HandoffConversationRequest) (*messengerv1.HandoffConversationResponse, error) {
	agentUserID, err := s.resolveActorUser(ctx, req.GetTenantId())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	ch, msg, _, err := s.svc.HandoffConversation(ctx, req.GetTenantId(), agentUserID, req.GetChannelId(), req.GetToUserId(), req.GetReason(), req.GetContextSummary())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return &messengerv1.HandoffConversationResponse{
		Channel:       toProtoChannel(ch),
		SystemMessage: toProtoMessage(msg),
	}, nil
}

func (s *MessengerServer) resolveActorUser(ctx context.Context, tenantID string) (string, error) {
	if uid, ok := MessengerUserIDFromContext(ctx); ok {
		return uid, nil
	}
	accountID, ok := AccountIDFromContext(ctx)
	if !ok {
		return "", messenger.ErrUnauthenticated
	}
	user, err := s.svc.ResolveHumanUser(ctx, tenantID, accountID, "")
	if err != nil {
		return "", err
	}
	return user.ID, nil
}
