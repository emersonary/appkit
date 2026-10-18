package vclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/emersonary/appkit/messenger"
	"github.com/emersonary/appkit/messenger/message"
	"github.com/emersonary/appkit/messenger/receipt"
	"github.com/emersonary/appkit/messenger/transport"
	"github.com/emersonary/appkit/messenger/ws"
	platformuuid "github.com/emersonary/appkit/uuid"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func TestGroupMessageIsHandledOnceAndClientsReceipt(t *testing.T) {
	nc, shutdown := startNATS(t)
	defer shutdown()

	var mu sync.Mutex
	publishes := 0
	saves := 0
	members := []string{"user-a", "user-b"}

	pub := messenger.NewPublisher(nc, nil)
	api := &socketAPI{
		onSend: func(ctx context.Context, actor, channelID, body, deviceID string) (messenger.Message, error) {
			mu.Lock()
			publishes++
			mu.Unlock()
			msg := messenger.Message{
				ID: platformuuid.NewString(), ChannelID: channelID, Body: &body, SenderUserID: &actor, CreatedAt: time.Now(),
			}
			err := message.PublishOnce(ctx, memberList{members}, natsMessagePub{pub}, message.Created{
				TenantID: "tenant", ChannelID: channelID, Message: msg, ExcludeDeviceID: deviceID,
			})
			return msg, err
		},
		onReceipt: func(ctx context.Context, actor, channelID, messageID, deviceID string) error {
			mu.Lock()
			saves++
			mu.Unlock()
			return receipt.Save(ctx, countingRepo{}, memberList{members}, natsReceiptPub{pub}, receipt.Record{
				ChannelID: channelID, MessageID: messageID, MessengerUserID: actor, DeviceID: deviceID,
			})
		},
	}

	hub := ws.NewHub()
	sub := messenger.NewSubscriber(nc, hub, nil)
	if err := sub.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Let the subscriptions land before clients publish.
	if err := nc.Flush(); err != nil {
		t.Fatal(err)
	}

	handler := transport.NewRealtimeHandler(api, hub, nil, "agent-secret")
	handler.PingInterval = time.Hour
	handler.PongTimeout = time.Hour

	srv := httptest.NewServer(http.HandlerFunc(handler.ServeHTTP))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	alice, err := Dial(ctx, srv.URL, "tenant", "user-a", "phone-a", "agent-secret")
	if err != nil {
		t.Fatal(err)
	}
	defer alice.Close()
	bob, err := Dial(ctx, srv.URL, "tenant", "user-b", "phone-b", "agent-secret", WithAutoReceipt())
	if err != nil {
		t.Fatal(err)
	}
	defer bob.Close()

	waitUntil(t, func() bool { return hub.Len() == 2 })

	if err := alice.SendMessage(ctx, "channel-1", "hello group"); err != nil {
		t.Fatal(err)
	}

	ev := waitEvent(t, bob.Events(), "message.created")
	var created messenger.MessageCreatedEvent
	if err := json.Unmarshal(ev.Payload, &created); err != nil {
		t.Fatal(err)
	}
	if created.Message.ID == "" || len(created.RecipientUserIDs) != 2 {
		t.Fatalf("event=%+v", created)
	}

	waitEvent(t, alice.Events(), "read.updated")

	mu.Lock()
	defer mu.Unlock()
	if publishes != 1 {
		t.Fatalf("group message published %d times", publishes)
	}
	if saves != 1 {
		t.Fatalf("read receipt saved %d times", saves)
	}
}

type memberList struct{ ids []string }

func (m memberList) ListMemberIDs(context.Context, string) ([]string, error) { return m.ids, nil }

type natsMessagePub struct{ p *messenger.Publisher }

func (n natsMessagePub) PublishCreated(ctx context.Context, event message.Created) error {
	msg, _ := event.Message.(messenger.Message)
	return n.p.PublishMessageCreated(ctx, event.TenantID, event.ChannelID, msg, event.ExcludeDeviceID, event.RecipientUserIDs)
}

type countingRepo struct{}

func (countingRepo) SaveReadReceipt(context.Context, string, string, string, string) error {
	return nil
}

type natsReceiptPub struct{ p *messenger.Publisher }

func (n natsReceiptPub) PublishRead(ctx context.Context, event receipt.Record) error {
	return n.p.PublishReadUpdated(ctx, event.ChannelID, event.MessengerUserID, event.MessageID, event.DeviceID, event.RecipientUserIDs)
}

type socketAPI struct {
	onSend    func(ctx context.Context, actor, channelID, body, deviceID string) (messenger.Message, error)
	onReceipt func(ctx context.Context, actor, channelID, messageID, deviceID string) error
}

func (s *socketAPI) RegisterDevice(_ context.Context, userID, deviceKey, platform, pushToken string) (messenger.Device, error) {
	return messenger.Device{ID: "device-" + deviceKey, MessengerUserID: userID, DeviceKey: deviceKey, Platform: platform}, nil
}

func (s *socketAPI) SendUserMessage(ctx context.Context, tenantID, actorUserID, channelID, body, clientRef, replyToID, deviceID string) (messenger.Message, error) {
	return s.onSend(ctx, actorUserID, channelID, body, deviceID)
}

func (s *socketAPI) RecordReadReceipt(ctx context.Context, tenantID, actorUserID, channelID, messageID, deviceID string) error {
	return s.onReceipt(ctx, actorUserID, channelID, messageID, deviceID)
}

func (s *socketAPI) GetUserByID(_ context.Context, tenantID, userID string) (messenger.User, error) {
	return messenger.User{ID: userID, TenantID: tenantID, UserType: messenger.UserTypeAgent, DisplayName: userID}, nil
}

func (s *socketAPI) ResolveHumanUser(context.Context, string, string, string) (messenger.User, error) {
	return messenger.User{}, messenger.ErrUnauthenticated
}

func startNATS(t *testing.T) (*nats.Conn, func()) {
	t.Helper()
	opts := &server.Options{Host: "127.0.0.1", Port: -1, NoLog: true, NoSigs: true}
	ns, err := server.NewServer(opts)
	if err != nil {
		t.Fatal(err)
	}
	go ns.Start()
	if !ns.ReadyForConnections(3 * time.Second) {
		t.Fatal("nats did not start")
	}
	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		ns.Shutdown()
		t.Fatal(err)
	}
	return nc, func() {
		nc.Close()
		ns.Shutdown()
	}
}

func waitEvent(t *testing.T, events <-chan Event, typ string) Event {
	t.Helper()
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	for {
		select {
		case ev, ok := <-events:
			if !ok {
				t.Fatal("event stream closed")
			}
			if ev.Type == typ {
				return ev
			}
		case <-timer.C:
			t.Fatalf("timed out waiting for %s", typ)
		}
	}
}

func waitUntil(t *testing.T, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met")
}
