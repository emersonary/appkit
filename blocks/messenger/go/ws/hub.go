package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// WriteTimeout bounds a single fan-out write. A stuck socket must not stall
// delivery to the other connections of the same event.
var WriteTimeout = 2 * time.Second

type Connection struct {
	MessengerUserID string
	DeviceID        string
	TenantID        string
	Conn            *websocket.Conn
}

type Hub struct {
	mu       sync.RWMutex
	byUser   map[string]map[*Connection]struct{}
	byDevice map[string]*Connection
}

func NewHub() *Hub {
	return &Hub{
		byUser:   make(map[string]map[*Connection]struct{}),
		byDevice: make(map[string]*Connection),
	}
}

// Register adds the connection. If this device already has a live socket,
// that previous connection is removed and returned so the caller can close it
// immediately. The server may not have noticed the old socket was dead yet.
func (h *Hub) Register(c *Connection) (replaced *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c.DeviceID != "" {
		if prev, ok := h.byDevice[c.DeviceID]; ok && prev != c {
			h.removeLocked(prev)
			replaced = prev
		}
		h.byDevice[c.DeviceID] = c
	}
	set := h.byUser[c.MessengerUserID]
	if set == nil {
		set = make(map[*Connection]struct{})
		h.byUser[c.MessengerUserID] = set
	}
	set[c] = struct{}{}
	return replaced
}

func (h *Hub) Unregister(c *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.removeLocked(c)
}

func (h *Hub) removeLocked(c *Connection) {
	if c == nil {
		return
	}
	if set := h.byUser[c.MessengerUserID]; set != nil {
		delete(set, c)
		if len(set) == 0 {
			delete(h.byUser, c.MessengerUserID)
		}
	}
	if c.DeviceID != "" && h.byDevice[c.DeviceID] == c {
		delete(h.byDevice, c.DeviceID)
	}
}

func (h *Hub) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	n := 0
	for _, set := range h.byUser {
		n += len(set)
	}
	return n
}

func (h *Hub) HasDevice(deviceID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.byDevice[deviceID]
	return ok
}

// BroadcastToUsers writes one event to the connections of the given users.
// Recipients are chosen by the service. This method does not look up membership.
func (h *Hub) BroadcastToUsers(userIDs []string, excludeDeviceID string, eventType string, payload any) {
	body, err := json.Marshal(map[string]any{"type": eventType, "payload": payload})
	if err != nil || len(userIDs) == 0 {
		return
	}

	seen := make(map[*Connection]struct{})
	var targets []*Connection
	h.mu.RLock()
	for _, uid := range userIDs {
		for conn := range h.byUser[uid] {
			if excludeDeviceID != "" && conn.DeviceID == excludeDeviceID {
				continue
			}
			if _, ok := seen[conn]; ok {
				continue
			}
			seen[conn] = struct{}{}
			targets = append(targets, conn)
		}
	}
	h.mu.RUnlock()

	var failed []*Connection
	for _, conn := range targets {
		if conn.Conn == nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), WriteTimeout)
		err := conn.Conn.Write(ctx, websocket.MessageText, body)
		cancel()
		if err != nil {
			failed = append(failed, conn)
		}
	}
	for _, conn := range failed {
		h.Unregister(conn)
		_ = conn.Conn.CloseNow()
	}
}
