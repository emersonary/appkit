package messenger

import (
	"context"

	"github.com/emersonary/appkit/messenger/channel"
	"github.com/emersonary/appkit/messenger/message"
	"github.com/emersonary/appkit/messenger/receipt"
)

type messagePublisher struct{ p *Publisher }

func (a messagePublisher) PublishCreated(ctx context.Context, event message.Created) error {
	msg, _ := event.Message.(Message)
	return a.p.PublishMessageCreated(ctx, event.TenantID, event.ChannelID, msg, event.ExcludeDeviceID, event.RecipientUserIDs)
}

type receiptRepo struct{ s *Store }

func (a receiptRepo) SaveReadReceipt(ctx context.Context, channelID, userID, messageID, deviceID string) error {
	_, err := a.s.SaveReadReceipt(ctx, channelID, userID, messageID, deviceID)
	return err
}

type receiptPublisher struct{ p *Publisher }

func (a receiptPublisher) PublishRead(ctx context.Context, event receipt.Record) error {
	return a.p.PublishReadUpdated(ctx, event.ChannelID, event.MessengerUserID, event.MessageID, event.DeviceID, event.RecipientUserIDs)
}

type channelPublisher struct{ p *Publisher }

func (a channelPublisher) PublishMembers(ctx context.Context, event channel.MembersUpdated) error {
	return a.p.PublishMemberUpdated(ctx, event.ChannelID, event.RecipientUserIDs)
}
