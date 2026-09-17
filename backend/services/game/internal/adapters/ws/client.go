package ws

import (
	"context"
	"shared/wsclient"
	"sync"
	"time"
)

type (
	onConn    func(ctx context.Context, clientID string) error
	onDisconn func(ctx context.Context, client *client) error
	onTimeout func(ctx context.Context, clientID string) error
)

type client struct {
	mu               sync.RWMutex
	id               string
	sessionID        string
	wsClient         *wsclient.Client
	isOnline         bool
	timeoutActive    bool
	connTimer        *time.Timer
	disconnExpiresAt time.Time
}

func newClient(id string) *client {
	return &client{
		id:            id,
		isOnline:      false,
		timeoutActive: false,
	}
}

func (c *client) ID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.id
}

func (c *client) SessionID() string {
	return c.sessionID
}

func (c *client) IsOnline() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.isOnline
}

func (c *client) TimeoutActive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.timeoutActive
}

func (c *client) DisconnExpiresAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.disconnExpiresAt
}

func (c *client) Connect(
	sessionID string,
	wsClient *wsclient.Client,
	onConn onConn,
) error {
	c.mu.Lock()
	c.sessionID = sessionID
	c.wsClient = wsClient
	c.isOnline = true
	c.timeoutActive = false
	c.disconnExpiresAt = time.Time{}
	if c.connTimer != nil {
		c.connTimer.Stop()
		c.connTimer = nil
	}
	clientID := c.id
	c.mu.Unlock()

	if onConn != nil {
		return onConn(context.Background(), clientID)
	}

	return nil
}

func (c *client) StartInitialTimeout(
	onTimeout onTimeout,
) {
	c.mu.Lock()
	c.wsClient = nil
	c.isOnline = false
	c.timeoutActive = true
	c.disconnExpiresAt = time.Now().UTC().Add(1 * time.Minute)

	if c.connTimer != nil {
		c.connTimer.Stop()
	}

	c.connTimer = c.initConnTimer(onTimeout)
	c.mu.Unlock()
}

func (c *client) initConnTimer(
	onTimeout onTimeout,
) *time.Timer {
	return time.AfterFunc(1*time.Minute, func() {
		if c.IsOnline() || !c.TimeoutActive() {
			return
		}

		if onTimeout != nil {
			_ = onTimeout(context.Background(), c.ID())
		}
	})
}

func (c *client) Disconnect(
	sessionID string,
	onDisconn onDisconn,
	onTimeout onTimeout,
) error {
	c.mu.Lock()

	if c.sessionID != sessionID {
		c.mu.Unlock()

		return nil
	}

	ws := c.wsClient
	c.wsClient = nil
	c.isOnline = false
	c.timeoutActive = true
	c.disconnExpiresAt = time.Now().UTC().Add(1 * time.Minute)

	if c.connTimer != nil {
		c.connTimer.Stop()
	}

	c.connTimer = c.initConnTimer(onTimeout)
	c.mu.Unlock()

	if ws != nil {
		ws.Close()
	}

	if onDisconn != nil {
		return onDisconn(context.Background(), c)
	}

	return nil
}

func (c *client) Send(ctx context.Context, msg []byte) error {
	c.mu.RLock()
	isOnline := c.isOnline
	ws := c.wsClient
	c.mu.RUnlock()

	if !isOnline || ws == nil {
		return nil
	}

	return ws.Send(ctx, msg)
}
