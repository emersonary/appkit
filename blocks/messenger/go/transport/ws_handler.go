package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/emersonary/appkit/messenger"
	"github.com/emersonary/appkit/messenger/ws"
)

const wsPath = "/messenger/ws"

// RealtimeAPI is the socket-facing slice of the messenger service.
type RealtimeAPI interface {
	RegisterDevice(ctx context.Context, userID, deviceKey, platform, pushToken string) (messenger.Device, error)
	SendUserMessage(ctx context.Context, tenantID, actorUserID, channelID, body, clientRef, replyToID, deviceID string) (messenger.Message, error)
	RecordReadReceipt(ctx context.Context, tenantID, actorUserID, channelID, messageID, deviceID string) error
	GetUserByID(ctx context.Context, tenantID, userID string) (messenger.User, error)
	ResolveHumanUser(ctx context.Context, tenantID, accountID, displayName string) (messenger.User, error)
}

type WSHandler struct {
	svc              RealtimeAPI
	hub              *ws.Hub
	resolveAccountID AccountIDResolver
	agentAPIKey      string
	// PingInterval and PongTimeout override the socket keepalive. Zero uses the defaults.
	PingInterval time.Duration
	PongTimeout  time.Duration
}

func NewWSHandler(block *messenger.Block, resolveAccountID AccountIDResolver, agentAPIKey string) *WSHandler {
	if block == nil {
		return nil
	}
	return NewRealtimeHandler(block.Service, block.Hub, resolveAccountID, agentAPIKey)
}

// NewRealtimeHandler builds the websocket endpoint from the socket-facing API.
func NewRealtimeHandler(api RealtimeAPI, hub *ws.Hub, resolveAccountID AccountIDResolver, agentAPIKey string) *WSHandler {
	return &WSHandler{
		svc:              api,
		hub:              hub,
		resolveAccountID: resolveAccountID,
		agentAPIKey:      agentAPIKey,
	}
}

func (h *WSHandler) Register(mux *http.ServeMux) {
	if h == nil || mux == nil {
		return
	}
	mux.HandleFunc("GET "+wsPath, h.ServeHTTP)
}

func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil || h.hub == nil {
		http.Error(w, "messenger unavailable", http.StatusServiceUnavailable)
		return
	}

	tenantID := strings.TrimSpace(r.URL.Query().Get("tenant_id"))
	deviceKey := strings.TrimSpace(r.URL.Query().Get("device_id"))
	if tenantID == "" || deviceKey == "" {
		http.Error(w, "tenant_id and device_id required", http.StatusBadRequest)
		return
	}

	userID, err := h.resolveMessengerUser(r.Context(), r, tenantID)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	dev, err := h.svc.RegisterDevice(r.Context(), userID, deviceKey, "web", "")
	if err != nil {
		http.Error(w, "device registration failed", http.StatusBadRequest)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}

	c := &ws.Connection{
		MessengerUserID: userID,
		DeviceID:        dev.ID,
		TenantID:        tenantID,
		Conn:            conn,
	}
	if prev := h.hub.Register(c); prev != nil && prev.Conn != nil {
		_ = prev.Conn.CloseNow()
	}
	defer func() {
		h.hub.Unregister(c)
		_ = conn.CloseNow()
	}()

	_ = ws.Run(r.Context(), conn, ws.SessionConfig{
		PingInterval: h.PingInterval,
		PongTimeout:  h.PongTimeout,
		OnFrame: func(ctx context.Context, data []byte) {
			h.handleFrame(ctx, c, data)
		},
	})
}

type clientFrame struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type sendPayload struct {
	ChannelID string `json:"channel_id"`
	Body      string `json:"body"`
	ClientRef string `json:"client_ref"`
	ReplyToID string `json:"reply_to_id"`
}

type readPayload struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
}

func (h *WSHandler) handleFrame(ctx context.Context, c *ws.Connection, data []byte) {
	var frame clientFrame
	if err := json.Unmarshal(data, &frame); err != nil || h.svc == nil {
		return
	}
	switch frame.Type {
	case "message.send":
		var p sendPayload
		if err := json.Unmarshal(frame.Payload, &p); err != nil || p.ChannelID == "" {
			return
		}
		_, _ = h.svc.SendUserMessage(ctx, c.TenantID, c.MessengerUserID, p.ChannelID, p.Body, p.ClientRef, p.ReplyToID, c.DeviceID)
	case "receipt.read":
		var p readPayload
		if err := json.Unmarshal(frame.Payload, &p); err != nil || p.ChannelID == "" || p.MessageID == "" {
			return
		}
		_ = h.svc.RecordReadReceipt(ctx, c.TenantID, c.MessengerUserID, p.ChannelID, p.MessageID, c.DeviceID)
	}
}

func (h *WSHandler) resolveMessengerUser(ctx context.Context, r *http.Request, tenantID string) (string, error) {
	agentKey := strings.TrimSpace(r.Header.Get("X-Messenger-Agent-Key"))
	if agentKey == "" {
		agentKey = strings.TrimSpace(r.URL.Query().Get("agent_key"))
	}
	agentUserID := strings.TrimSpace(r.Header.Get("X-Messenger-Agent-User-Id"))
	if agentUserID == "" {
		agentUserID = strings.TrimSpace(r.URL.Query().Get("agent_user_id"))
	}
	if h.agentAPIKey != "" && agentKey == h.agentAPIKey && agentUserID != "" {
		user, err := h.svc.GetUserByID(ctx, tenantID, agentUserID)
		if err != nil {
			return "", err
		}
		if user.UserType != messenger.UserTypeAgent {
			return "", messenger.ErrForbidden
		}
		return user.ID, nil
	}

	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		}
	}
	if token == "" || h.resolveAccountID == nil {
		return "", messenger.ErrUnauthenticated
	}
	accountID, err := h.resolveAccountID(ctx, token)
	if err != nil || accountID == "" {
		return "", messenger.ErrUnauthenticated
	}
	user, err := h.svc.ResolveHumanUser(ctx, tenantID, accountID, "")
	if err != nil {
		return "", err
	}
	return user.ID, nil
}
