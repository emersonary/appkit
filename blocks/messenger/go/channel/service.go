// Package channel is the channel subject. Membership changes are resolved once
// here. The socket layer delivers the event to the listed connections.
package channel

import "context"

// MembersUpdated is one membership change for a channel.
type MembersUpdated struct {
	ChannelID        string   `json:"channel_id"`
	RecipientUserIDs []string `json:"recipient_user_ids"`
}

// Members lists the channel once.
type Members interface {
	ListMemberIDs(ctx context.Context, channelID string) ([]string, error)
}

// Publisher emits the channel subject.
type Publisher interface {
	PublishMembers(ctx context.Context, event MembersUpdated) error
}

// PublishMembersOnce resolves recipients a single time and publishes one event.
func PublishMembersOnce(ctx context.Context, members Members, publisher Publisher, channelID string) error {
	if publisher == nil {
		return nil
	}
	ids, err := members.ListMemberIDs(ctx, channelID)
	if err != nil {
		return err
	}
	return publisher.PublishMembers(ctx, MembersUpdated{ChannelID: channelID, RecipientUserIDs: ids})
}
