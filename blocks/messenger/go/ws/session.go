package ws

import (
	"context"
	"time"

	"github.com/coder/websocket"
)

const (
	// DefaultPingInterval is how often the server pings a live socket.
	// A dead connection is not noticed until a ping goes unanswered.
	DefaultPingInterval = 54 * time.Second
	// DefaultPongTimeout is how long the server waits for the pong.
	DefaultPongTimeout = 60 * time.Second
)

// SessionConfig controls the ping loop and inbound client frames.
type SessionConfig struct {
	PingInterval time.Duration
	PongTimeout  time.Duration
	OnFrame      func(ctx context.Context, data []byte)
}

func (c SessionConfig) pingInterval() time.Duration {
	if c.PingInterval > 0 {
		return c.PingInterval
	}
	return DefaultPingInterval
}

func (c SessionConfig) pongTimeout() time.Duration {
	if c.PongTimeout > 0 {
		return c.PongTimeout
	}
	return DefaultPongTimeout
}

// Run reads client frames and pings the peer until the socket dies.
// Ping failure, a replaced connection, or a dropped peer all end the session.
func Run(ctx context.Context, conn *websocket.Conn, cfg SessionConfig) error {
	if conn == nil {
		return nil
	}
	conn.SetReadLimit(1 << 20)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pingDone := make(chan struct{})
	go func() {
		defer close(pingDone)
		pingLoop(ctx, conn, cfg.pingInterval(), cfg.pongTimeout())
	}()

	err := readLoop(ctx, conn, cfg.OnFrame)
	cancel()
	<-pingDone
	return err
}

func readLoop(ctx context.Context, conn *websocket.Conn, onFrame func(context.Context, []byte)) error {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return err
		}
		if onFrame != nil && len(data) > 0 {
			onFrame(ctx, data)
		}
	}
}

func pingLoop(ctx context.Context, conn *websocket.Conn, interval, timeout time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pctx, cancel := context.WithTimeout(ctx, timeout)
			err := conn.Ping(pctx)
			cancel()
			if err != nil {
				// CloseNow: a dead peer will not complete a close handshake,
				// and waiting for one holds the socket until the timeout.
				_ = conn.CloseNow()
				return
			}
		}
	}
}
