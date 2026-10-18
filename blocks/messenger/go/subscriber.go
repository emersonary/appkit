package messenger

import (
	"context"
	"encoding/json"

	"github.com/emersonary/appkit/messenger/ws"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Subscriber is the socket side of NATS. It delivers events to connections.
// Recipients are already resolved by the service that published the event.
type Subscriber struct {
	nc     *nats.Conn
	hub    *ws.Hub
	logger *zap.Logger
}

func NewSubscriber(nc *nats.Conn, hub *ws.Hub, logger *zap.Logger) *Subscriber {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Subscriber{nc: nc, hub: hub, logger: logger.Named("subscriber")}
}

func (s *Subscriber) Start(ctx context.Context) error {
	if s == nil || s.nc == nil || s.hub == nil {
		return nil
	}

	_, err := s.nc.Subscribe("messenger.message.created.*", func(msg *nats.Msg) {
		var ev MessageCreatedEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			s.logger.Warn("invalid message created event", zap.Error(err))
			return
		}
		s.hub.BroadcastToUsers(ev.RecipientUserIDs, ev.ExcludeDeviceID, "message.created", ev)
	})
	if err != nil {
		return err
	}

	_, err = s.nc.Subscribe("messenger.read.updated.*.*", func(msg *nats.Msg) {
		var ev ReadUpdatedEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			return
		}
		ids := ev.RecipientUserIDs
		if len(ids) == 0 {
			ids = []string{ev.MessengerUserID}
		}
		s.hub.BroadcastToUsers(ids, "", "read.updated", ev)
	})
	if err != nil {
		return err
	}

	_, err = s.nc.Subscribe("messenger.member.updated.*", func(msg *nats.Msg) {
		var ev MemberUpdatedEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			return
		}
		s.hub.BroadcastToUsers(ev.RecipientUserIDs, "", "member.updated", ev)
	})
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		_ = s.nc.Drain()
	}()

	s.logger.Info("messenger NATS subscriber started")
	return nil
}
