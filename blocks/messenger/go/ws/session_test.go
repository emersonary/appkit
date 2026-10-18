package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestPingTimeoutDropsSilentPeer(t *testing.T) {
	hub := NewHub()
	accepted := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		c := &Connection{MessengerUserID: "u1", DeviceID: "d1", Conn: conn}
		hub.Register(c)
		close(accepted)
		_ = Run(r.Context(), conn, SessionConfig{
			PingInterval: 20 * time.Millisecond,
			PongTimeout:  40 * time.Millisecond,
		})
		hub.Unregister(c)
		_ = conn.Close(websocket.StatusGoingAway, "dead")
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client, _, err := websocket.Dial(ctx, "ws"+srv.URL[len("http"):], nil)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close(websocket.StatusNormalClosure, "done")

	select {
	case <-accepted:
	case <-ctx.Done():
		t.Fatal("server did not accept")
	}
	// Stop reading so pongs are never sent. The server must notice via ping.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if !hub.HasDevice("d1") {
			return
		}
		time.Sleep(15 * time.Millisecond)
	}
	t.Fatal("silent connection was still registered after ping timeout")
}

func TestReplacementClosesPreviousDeviceConnection(t *testing.T) {
	hub := NewHub()
	firstClosed := make(chan struct{})
	var order sync.Mutex
	n := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		order.Lock()
		n++
		mine := n
		order.Unlock()
		c := &Connection{MessengerUserID: "u1", DeviceID: "phone", Conn: conn}
		if prev := hub.Register(c); prev != nil && prev.Conn != nil {
			_ = prev.Conn.CloseNow()
		}
		_ = Run(r.Context(), conn, SessionConfig{PingInterval: time.Hour, PongTimeout: time.Hour})
		hub.Unregister(c)
		if mine == 1 {
			close(firstClosed)
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	url := "ws" + srv.URL[len("http"):]
	first, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close(websocket.StatusNormalClosure, "done")
	go discard(first)

	waitUntil(t, func() bool { return hub.HasDevice("phone") })

	second, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close(websocket.StatusNormalClosure, "done")
	go discard(second)

	select {
	case <-firstClosed:
	case <-ctx.Done():
		t.Fatal("first connection was not closed when the device reconnected")
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if hub.Len() == 1 && hub.HasDevice("phone") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("after replace len=%d has=%v", hub.Len(), hub.HasDevice("phone"))
}

func TestClientCloseUnregisters(t *testing.T) {
	hub := NewHub()
	ended := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		c := &Connection{MessengerUserID: "u1", DeviceID: "d1", Conn: conn}
		hub.Register(c)
		_ = Run(r.Context(), conn, SessionConfig{PingInterval: time.Hour, PongTimeout: time.Hour})
		hub.Unregister(c)
		close(ended)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client, _, err := websocket.Dial(ctx, "ws"+srv.URL[len("http"):], nil)
	if err != nil {
		t.Fatal(err)
	}
	go discard(client)
	waitUntil(t, func() bool { return hub.HasDevice("d1") })
	_ = client.Close(websocket.StatusNormalClosure, "bye")
	select {
	case <-ended:
	case <-ctx.Done():
		t.Fatal("server session did not end after the client closed")
	}
	if hub.HasDevice("d1") {
		t.Fatal("closed device still registered")
	}
}

func discard(conn *websocket.Conn) {
	for {
		if _, _, err := conn.Read(context.Background()); err != nil {
			return
		}
	}
}

func waitUntil(t *testing.T, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met")
}
