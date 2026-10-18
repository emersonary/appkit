package messenger

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/emersonary/appkit/messenger/channel"
	"github.com/emersonary/appkit/messenger/message"
	"github.com/emersonary/appkit/messenger/receipt"
)

type Service struct {
	store     *Store
	cfg       Config
	publisher *Publisher
	logger    *zap.Logger
}

func NewService(store *Store, cfg Config, publisher *Publisher, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		store:     store,
		cfg:       cfg,
		publisher: publisher,
		logger:    logger.Named("service"),
	}
}

func (s *Service) Config() Config { return s.cfg }

func (s *Service) ResolveHumanUser(ctx context.Context, tenantID, accountID, displayName string) (User, error) {
	if tenantID == "" || accountID == "" {
		return User{}, ErrInvalidInput.With("reason", "tenant_id and account_id required")
	}
	return s.store.UpsertHumanUser(ctx, tenantID, accountID, displayName)
}

func (s *Service) CreateAgentUser(ctx context.Context, tenantID, displayName, externalRef, avatarURL string) (User, error) {
	if tenantID == "" || displayName == "" {
		return User{}, ErrInvalidInput.With("reason", "tenant_id and display_name required")
	}
	if externalRef != "" {
		if existing, err := s.store.GetUserByExternalRef(ctx, tenantID, externalRef); err == nil {
			return existing, nil
		} else if !isNotFound(err) {
			return User{}, err
		}
	}
	return s.store.CreateAgentUser(ctx, tenantID, displayName, externalRef, avatarURL)
}

func (s *Service) GetAgentUser(ctx context.Context, tenantID, externalRef string) (User, error) {
	return s.store.GetUserByExternalRef(ctx, tenantID, externalRef)
}

func (s *Service) GetUserByID(ctx context.Context, tenantID, userID string) (User, error) {
	return s.store.GetUserByID(ctx, tenantID, userID)
}

func (s *Service) RegisterDevice(ctx context.Context, userID, deviceKey, platform, pushToken string) (Device, error) {
	if userID == "" || deviceKey == "" {
		return Device{}, ErrInvalidInput.With("reason", "user and device_key required")
	}
	return s.store.RegisterDevice(ctx, userID, deviceKey, platform, pushToken)
}

func (s *Service) GetOrCreateDirect(ctx context.Context, tenantID, actorUserID, otherUserID string) (Channel, error) {
	if actorUserID == otherUserID {
		return Channel{}, ErrInvalidInput.With("reason", "cannot create direct channel with self")
	}
	if _, err := s.store.GetUserByID(ctx, tenantID, otherUserID); err != nil {
		return Channel{}, err
	}
	return s.store.GetOrCreateDirect(ctx, tenantID, actorUserID, otherUserID, actorUserID)
}

func (s *Service) CreateGroup(ctx context.Context, tenantID, actorUserID, title string, memberUserIDs []string) (Channel, error) {
	if title == "" {
		return Channel{}, ErrInvalidInput.With("reason", "title required")
	}
	return s.store.CreateGroup(ctx, tenantID, title, actorUserID, memberUserIDs)
}

func (s *Service) ListChannels(ctx context.Context, tenantID, actorUserID string) ([]Channel, error) {
	return s.store.ListChannelsForUser(ctx, tenantID, actorUserID)
}

func (s *Service) ListMessages(ctx context.Context, tenantID, actorUserID, channelID, beforeID string, limit int) ([]Message, error) {
	if err := s.requireMember(ctx, channelID, actorUserID); err != nil {
		return nil, err
	}
	msgs, err := s.store.ListMessages(ctx, channelID, beforeID, limit)
	if err != nil {
		return nil, err
	}
	reverseMessages(msgs)
	return msgs, nil
}

func (s *Service) SendUserMessage(ctx context.Context, tenantID, actorUserID, channelID, body, clientRef, replyToID, deviceID string) (Message, error) {
	if body == "" {
		return Message{}, ErrInvalidInput.With("reason", "body required")
	}
	if err := s.requireMember(ctx, channelID, actorUserID); err != nil {
		return Message{}, err
	}
	if deviceID != "" {
		if _, err := s.store.GetDevice(ctx, deviceID, actorUserID); err != nil {
			return Message{}, err
		}
	}

	var clientRefPtr *string
	if clientRef != "" {
		clientRefPtr = &clientRef
	}
	sender := actorUserID
	msg, err := s.store.InsertMessage(ctx, InsertMessageInput{
		ChannelID: channelID, Kind: MessageKindUser, SenderUserID: &sender,
		Body: &body, ClientRef: clientRefPtr, ReplyToID: optionalString(replyToID),
	})
	if err != nil {
		return Message{}, err
	}
	s.publishMessageCreated(ctx, tenantID, channelID, msg, deviceID)
	return msg, nil
}

func (s *Service) SendSystemMessage(ctx context.Context, tenantID, channelID, eventType, body string, payload json.RawMessage, senderUserID *string) (Message, error) {
	if eventType == "" {
		return Message{}, ErrInvalidInput.With("reason", "event_type required")
	}
	_ = tenantID

	var bodyPtr *string
	if body != "" {
		bodyPtr = &body
	}
	var eventTypePtr = &eventType
	msg, err := s.store.InsertMessage(ctx, InsertMessageInput{
		ChannelID: channelID, Kind: MessageKindSystem, SenderUserID: senderUserID,
		Body: bodyPtr, EventType: eventTypePtr, EventPayload: payload,
	})
	if err != nil {
		return Message{}, err
	}
	s.publishMessageCreated(ctx, tenantID, channelID, msg, "")
	return msg, nil
}

func (s *Service) MarkRead(ctx context.Context, tenantID, actorUserID, channelID, messageID string) error {
	return s.RecordReadReceipt(ctx, tenantID, actorUserID, channelID, messageID, "")
}

// RecordReadReceipt stores one read, once, then publishes it for the channel's connections.
func (s *Service) RecordReadReceipt(ctx context.Context, tenantID, actorUserID, channelID, messageID, deviceID string) error {
	if err := s.requireMember(ctx, channelID, actorUserID); err != nil {
		return err
	}
	var publisher receipt.Publisher
	if s.publisher != nil {
		publisher = receiptPublisher{s.publisher}
	}
	if err := receipt.Save(ctx, receiptRepo{s.store}, s.store, publisher, receipt.Record{
		ChannelID:       channelID,
		MessageID:       messageID,
		MessengerUserID: actorUserID,
		DeviceID:        deviceID,
	}); err != nil {
		return err
	}
	return nil
}

type SyncResult struct {
	Channel  Channel
	Messages []Message
}

func (s *Service) SyncMessages(ctx context.Context, tenantID, actorUserID string, cursors map[string]string) ([]SyncResult, error) {
	if len(cursors) == 0 {
		channels, err := s.store.ListChannelsForUser(ctx, tenantID, actorUserID)
		if err != nil {
			return nil, err
		}
		cursors = make(map[string]string, len(channels))
		for _, ch := range channels {
			cursors[ch.ID] = ""
		}
	}

	out := make([]SyncResult, 0, len(cursors))
	for channelID, afterID := range cursors {
		if err := s.requireMember(ctx, channelID, actorUserID); err != nil {
			continue
		}
		msgs, err := s.store.MessagesAfter(ctx, channelID, afterID)
		if err != nil {
			return nil, err
		}
		ch, err := s.store.GetChannelForUser(ctx, tenantID, channelID, actorUserID)
		if err != nil {
			return nil, err
		}
		out = append(out, SyncResult{Channel: ch, Messages: msgs})
	}
	return out, nil
}

func (s *Service) HandoffConversation(ctx context.Context, tenantID, agentUserID, channelID, toUserID, reason, summary string) (Channel, Message, Handoff, error) {
	agent, err := s.store.GetUserByID(ctx, tenantID, agentUserID)
	if err != nil {
		return Channel{}, Message{}, Handoff{}, err
	}
	if agent.UserType != UserTypeAgent {
		return Channel{}, Message{}, Handoff{}, ErrForbidden.With("reason", "only agents can hand off")
	}
	if err := s.requireMember(ctx, channelID, agentUserID); err != nil {
		return Channel{}, Message{}, Handoff{}, err
	}
	if _, err := s.store.GetUserByID(ctx, tenantID, toUserID); err != nil {
		return Channel{}, Message{}, Handoff{}, err
	}

	if err := s.store.AddChannelMember(ctx, channelID, toUserID, MemberRoleMember); err != nil {
		return Channel{}, Message{}, Handoff{}, err
	}

	payload, _ := json.Marshal(map[string]string{
		"from_user_id": agentUserID,
		"to_user_id":   toUserID,
		"reason":       reason,
		"summary":      summary,
	})
	body := "Conversation transferred"
	if summary != "" {
		body = summary
	}
	msg, err := s.SendSystemMessage(ctx, tenantID, channelID, "handoff", body, payload, &agentUserID)
	if err != nil {
		return Channel{}, Message{}, Handoff{}, err
	}

	h, err := s.store.InsertHandoff(ctx, channelID, agentUserID, toUserID, optionalString(reason), optionalString(summary))
	if err != nil {
		return Channel{}, Message{}, Handoff{}, err
	}

	ch, err := s.store.GetChannelForUser(ctx, tenantID, channelID, agentUserID)
	if err != nil {
		return Channel{}, Message{}, Handoff{}, err
	}

	if s.publisher != nil {
		_ = channel.PublishMembersOnce(ctx, s.store, channelPublisher{s.publisher}, channelID)
	}
	return ch, msg, h, nil
}

func (s *Service) requireMember(ctx context.Context, channelID, userID string) error {
	ok, err := s.store.IsMember(ctx, channelID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func (s *Service) publishMessageCreated(ctx context.Context, tenantID, channelID string, msg Message, excludeDeviceID string) {
	if s.publisher == nil {
		return
	}
	if err := message.PublishOnce(ctx, s.store, messagePublisher{s.publisher}, message.Created{
		TenantID:        tenantID,
		ChannelID:       channelID,
		Message:         msg,
		ExcludeDeviceID: excludeDeviceID,
	}); err != nil {
		s.logger.Warn("publish message created failed", zap.Error(err))
	}
}

func optionalString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func reverseMessages(msgs []Message) {
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
}
