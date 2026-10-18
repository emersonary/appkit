// Package message is the message subject. A group message is handled once
// here. The socket layer only delivers the resulting event to connections.
package message

import "context"

// Created is published once per message, with every recipient already resolved.
type Created struct {
	TenantID         string   `json:"tenant_id"`
	ChannelID        string   `json:"channel_id"`
	Message          any      `json:"message"`
	ExcludeDeviceID  string   `json:"exclude_device_id,omitempty"`
	RecipientUserIDs []string `json:"recipient_user_ids"`
}

// Members lists the channel once, in the service, before publish.
type Members interface {
	ListMemberIDs(ctx context.Context, channelID string) ([]string, error)
}

// Publisher sends the event on the message subject.
type Publisher interface {
	PublishCreated(ctx context.Context, event Created) error
}

// PublishOnce resolves recipients a single time and publishes one event.
func PublishOnce(ctx context.Context, members Members, publisher Publisher, event Created) error {
	if publisher == nil {
		return nil
	}
	ids, err := members.ListMemberIDs(ctx, event.ChannelID)
	if err != nil {
		return err
	}
	event.RecipientUserIDs = ids
	return publisher.PublishCreated(ctx, event)
}
