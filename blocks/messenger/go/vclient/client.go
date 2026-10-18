// Package vclient is a Go messenger client used by end-to-end tests.
// It speaks the same websocket frames as a device: send a message, and send a
// read receipt after a message arrives.
package vclient

import (
	"context"
	"encoding/json"
	"net/url"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// Event is one server frame.
type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Client is one device connection.
type Client struct {
	UserID    string
	DeviceKey string

	conn        *websocket.Conn
	events      chan Event
	autoReceipt bool
	writeMu     sync.Mutex
	done        chan struct{}
	once        sync.Once
}

type Option func(*Client)

// WithAutoReceipt sends receipt.read for every message.created that is not our own.
func WithAutoReceipt() Option {
	return func(c *Client) { c.autoReceipt = true }
}

// Dial opens a websocket as an agent device. baseURL is the http(s) origin, without a path.
func Dial(ctx context.Context, baseURL, tenantID, userID, deviceKey, agentKey string, opts ...Option) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	default:
		u.Scheme = "ws"
	}
	q := u.Query()
	q.Set("tenant_id", tenantID)
	q.Set("device_id", deviceKey)
	q.Set("agent_key", agentKey)
	q.Set("agent_user_id", userID)
	u.RawQuery = q.Encode()
	u.Path = "/messenger/ws"

	conn, _, err := websocket.Dial(ctx, u.String(), nil)
	if err != nil {
		return nil, err
	}
	c := &Client{
		UserID:    userID,
		DeviceKey: deviceKey,
		conn:      conn,
		events:    make(chan Event, 32),
		done:      make(chan struct{}),
	}
	for _, opt := range opts {
		opt(c)
	}
	go c.readLoop()
	return c, nil
}

func (c *Client) Events() <-chan Event { return c.events }

func (c *Client) Close() error {
	var err error
	c.once.Do(func() {
		err = c.conn.Close(websocket.StatusNormalClosure, "client closed")
		close(c.done)
	})
	return err
}

// SendMessage sends a user message on a channel.
func (c *Client) SendMessage(ctx context.Context, channelID, body string) error {
	return c.write(ctx, "message.send", map[string]string{
		"channel_id": channelID,
		"body":       body,
	})
}

// SendReadReceipt reports that this device has read a message.
func (c *Client) SendReadReceipt(ctx context.Context, channelID, messageID string) error {
	return c.write(ctx, "receipt.read", map[string]string{
		"channel_id": channelID,
		"message_id": messageID,
	})
}

func (c *Client) write(ctx context.Context, typ string, payload any) error {
	body, err := json.Marshal(map[string]any{"type": typ, "payload": payload})
	if err != nil {
		return err
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.Write(ctx, websocket.MessageText, body)
}

func (c *Client) readLoop() {
	defer close(c.events)
	for {
		_, data, err := c.conn.Read(context.Background())
		if err != nil {
			return
		}
		var ev Event
		if err := json.Unmarshal(data, &ev); err != nil {
			continue
		}
		if c.autoReceipt && ev.Type == "message.created" {
			c.ack(ev)
		}
		select {
		case c.events <- ev:
		case <-c.done:
			return
		}
	}
}

func (c *Client) ack(ev Event) {
	var payload struct {
		Message struct {
			ID           string  `json:"ID"`
			ChannelID    string  `json:"ChannelID"`
			SenderUserID *string `json:"SenderUserID"`
		} `json:"message"`
	}
	if err := json.Unmarshal(ev.Payload, &payload); err != nil || payload.Message.ID == "" {
		return
	}
	if payload.Message.SenderUserID != nil && *payload.Message.SenderUserID == c.UserID {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = c.SendReadReceipt(ctx, payload.Message.ChannelID, payload.Message.ID)
}
