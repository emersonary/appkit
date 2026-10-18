package transport

import (
	"encoding/json"

	messengerv1 "github.com/emersonary/appkit/messenger/gen/messenger/v1"
	"github.com/emersonary/appkit/messenger"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoUser(u messenger.User) *messengerv1.MessengerUser {
	out := &messengerv1.MessengerUser{
		Id:          u.ID,
		TenantId:    u.TenantID,
		UserType:    toProtoUserType(u.UserType),
		DisplayName: u.DisplayName,
	}
	if u.AvatarURL != nil {
		out.AvatarUrl = *u.AvatarURL
	}
	if u.AccountID != nil {
		out.AccountId = *u.AccountID
	}
	if u.ExternalRef != nil {
		out.ExternalRef = *u.ExternalRef
	}
	return out
}

func toProtoUserType(t string) messengerv1.UserType {
	switch t {
	case messenger.UserTypeHuman:
		return messengerv1.UserType_USER_TYPE_HUMAN
	case messenger.UserTypeAgent:
		return messengerv1.UserType_USER_TYPE_AGENT
	case messenger.UserTypeSystem:
		return messengerv1.UserType_USER_TYPE_SYSTEM
	default:
		return messengerv1.UserType_USER_TYPE_UNSPECIFIED
	}
}

func toProtoChannel(ch messenger.Channel) *messengerv1.Channel {
	out := &messengerv1.Channel{
		Id:              ch.ID,
		TenantId:        ch.TenantID,
		Kind:            toProtoChannelKind(ch.Kind),
		CreatedByUserId: ch.CreatedByUserID,
		UnreadCount:     ch.UnreadCount,
		UpdatedAt:       timestamppb.New(ch.UpdatedAt),
	}
	if ch.Title != nil {
		out.Title = *ch.Title
	}
	if ch.AvatarURL != nil {
		out.AvatarUrl = *ch.AvatarURL
	}
	if ch.LastPreview != nil {
		out.LastMessagePreview = *ch.LastPreview
	}
	if len(ch.Members) > 0 {
		out.Members = make([]*messengerv1.MessengerUser, 0, len(ch.Members))
		for _, m := range ch.Members {
			out.Members = append(out.Members, toProtoUser(m.User))
		}
	}
	return out
}

func toProtoChannelKind(k string) messengerv1.ChannelKind {
	switch k {
	case messenger.ChannelKindDirect:
		return messengerv1.ChannelKind_CHANNEL_KIND_DIRECT
	case messenger.ChannelKindGroup:
		return messengerv1.ChannelKind_CHANNEL_KIND_GROUP
	default:
		return messengerv1.ChannelKind_CHANNEL_KIND_UNSPECIFIED
	}
}

func toProtoMessage(msg messenger.Message) *messengerv1.Message {
	out := &messengerv1.Message{
		Id:        msg.ID,
		ChannelId: msg.ChannelID,
		Kind:      toProtoMessageKind(msg.Kind),
		CreatedAt: timestamppb.New(msg.CreatedAt),
	}
	if msg.Body != nil {
		out.Body = *msg.Body
	}
	if msg.EventType != nil {
		out.EventType = *msg.EventType
	}
	if len(msg.EventPayload) > 0 {
		var m map[string]any
		if err := json.Unmarshal(msg.EventPayload, &m); err == nil {
			if st, err := structpb.NewStruct(m); err == nil {
				out.EventPayload = st
			}
		}
	}
	if msg.ClientRef != nil {
		out.ClientRef = *msg.ClientRef
	}
	if msg.ReplyToID != nil {
		out.ReplyToId = *msg.ReplyToID
	}
	if msg.Sender != nil {
		out.Sender = toProtoUser(*msg.Sender)
	}
	return out
}

func toProtoMessageKind(k string) messengerv1.MessageKind {
	switch k {
	case messenger.MessageKindUser:
		return messengerv1.MessageKind_MESSAGE_KIND_USER
	case messenger.MessageKindSystem:
		return messengerv1.MessageKind_MESSAGE_KIND_SYSTEM
	default:
		return messengerv1.MessageKind_MESSAGE_KIND_UNSPECIFIED
	}
}

func structToRaw(st *structpb.Struct) json.RawMessage {
	if st == nil {
		return nil
	}
	b, err := json.Marshal(st.AsMap())
	if err != nil {
		return nil
	}
	return json.RawMessage(b)
}
