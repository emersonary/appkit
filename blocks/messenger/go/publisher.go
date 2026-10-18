package messenger

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Publisher struct {
	nc     *nats.Conn
	logger *zap.Logger
}

func NewPublisher(nc *nats.Conn, logger *zap.Logger) *Publisher {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Publisher{nc: nc, logger: logger.Named("publisher")}
}

type MessageCreatedEvent struct {
	TenantID         string   `json:"tenant_id"`
	ChannelID        string   `json:"channel_id"`
	Message          Message  `json:"message"`
	ExcludeDeviceID  string   `json:"exclude_device_id,omitempty"`
	RecipientUserIDs []string `json:"recipient_user_ids"`
}

type ReadUpdatedEvent struct {
	ChannelID        string   `json:"channel_id"`
	MessengerUserID  string   `json:"messenger_user_id"`
	MessageID        string   `json:"message_id"`
	DeviceID         string   `json:"device_id,omitempty"`
	RecipientUserIDs []string `json:"recipient_user_ids"`
}

type MemberUpdatedEvent struct {
	ChannelID        string   `json:"channel_id"`
	RecipientUserIDs []string `json:"recipient_user_ids"`
}

func (p *Publisher) PublishMessageCreated(ctx context.Context, tenantID, channelID string, msg Message, excludeDeviceID string, recipientUserIDs []string) error {
	if p == nil || p.nc == nil {
		return nil
	}
	body, err := json.Marshal(MessageCreatedEvent{
		TenantID: tenantID, ChannelID: channelID, Message: msg, ExcludeDeviceID: excludeDeviceID, RecipientUserIDs: recipientUserIDs,
	})
	if err != nil {
		return err
	}
	subject := fmt.Sprintf("messenger.message.created.%s", channelID)
	return p.nc.Publish(subject, body)
}

func (p *Publisher) PublishReadUpdated(ctx context.Context, channelID, messengerUserID, messageID, deviceID string, recipientUserIDs []string) error {
	if p == nil || p.nc == nil {
		return nil
	}
	body, err := json.Marshal(ReadUpdatedEvent{
		ChannelID: channelID, MessengerUserID: messengerUserID, MessageID: messageID, DeviceID: deviceID, RecipientUserIDs: recipientUserIDs,
	})
	if err != nil {
		return err
	}
	subject := fmt.Sprintf("messenger.read.updated.%s.%s", channelID, messengerUserID)
	return p.nc.Publish(subject, body)
}

func (p *Publisher) PublishMemberUpdated(ctx context.Context, channelID string, recipientUserIDs []string) error {
	if p == nil || p.nc == nil {
		return nil
	}
	body, err := json.Marshal(MemberUpdatedEvent{ChannelID: channelID, RecipientUserIDs: recipientUserIDs})
	if err != nil {
		return err
	}
	subject := fmt.Sprintf("messenger.member.updated.%s", channelID)
	return p.nc.Publish(subject, body)
}
