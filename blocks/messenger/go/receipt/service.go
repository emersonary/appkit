// Package receipt is the read-receipt subject. Clients report a read after
// they receive a message. This service stores it once, then publishes one event.
package receipt

import "context"

// Record is one user's read of one message.
type Record struct {
	ChannelID        string   `json:"channel_id"`
	MessageID        string   `json:"message_id"`
	MessengerUserID  string   `json:"messenger_user_id"`
	DeviceID         string   `json:"device_id,omitempty"`
	RecipientUserIDs []string `json:"recipient_user_ids"`
}

// Repo persists the receipt.
type Repo interface {
	SaveReadReceipt(ctx context.Context, channelID, userID, messageID, deviceID string) error
}

// Members resolves who else should hear about the receipt. Called once.
type Members interface {
	ListMemberIDs(ctx context.Context, channelID string) ([]string, error)
}

// Publisher emits the receipt on its NATS subject.
type Publisher interface {
	PublishRead(ctx context.Context, event Record) error
}

// Save stores the receipt and publishes one event for the channel.
func Save(ctx context.Context, repo Repo, members Members, publisher Publisher, event Record) error {
	if err := repo.SaveReadReceipt(ctx, event.ChannelID, event.MessengerUserID, event.MessageID, event.DeviceID); err != nil {
		return err
	}
	if publisher == nil {
		return nil
	}
	ids, err := members.ListMemberIDs(ctx, event.ChannelID)
	if err != nil {
		return err
	}
	event.RecipientUserIDs = ids
	return publisher.PublishRead(ctx, event)
}
