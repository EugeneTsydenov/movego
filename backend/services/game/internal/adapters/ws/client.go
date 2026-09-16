package ws

import (
	"context"
	"shared/wsclient"
	"sync"
	"time"
)

type client struct {
	mu               sync.RWMutex
	clientID         string
	wsClient         *wsclient.Client
	isOnline         bool
	timeoutActive    bool
	connTimer        *time.Timer
	disconnExpiresAt time.Time
}

func newOnlineClient(clientID string, wsClient *wsclient.Client) *client {
	return &client{
		clientID:      clientID,
		wsClient:      wsClient,
		isOnline:      true,
		timeoutActive: false,
	}
}

func newOfflineClient(clientID string) *client {
	return &client{
		clientID:      clientID,
		isOnline:      false,
		timeoutActive: false,
	}
}

func (c *client) ClientID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.clientID
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

func (c *client) Connect(wsClient *wsclient.Client, onConn func(ctx context.Context, clientID string) error) error {
	c.mu.Lock()
	c.wsClient = wsClient
	c.isOnline = true
	c.timeoutActive = false
	c.disconnExpiresAt = time.Time{}
	if c.connTimer != nil {
		c.connTimer.Stop()
		c.connTimer = nil
	}
	clientID := c.clientID
	c.mu.Unlock()

	if onConn != nil {
		return onConn(context.Background(), clientID)
	}
	return nil
}

func (c *client) Disconnect(
	onDisconn func(ctx context.Context, client *client) error,
	onTimeout func(ctx context.Context, clientID string) error,
) error {
	c.mu.Lock()
	ws := c.wsClient
	c.wsClient = nil
	c.isOnline = false
	c.timeoutActive = true
	c.disconnExpiresAt = time.Now().UTC().Add(1 * time.Minute)

	if c.connTimer != nil {
		c.connTimer.Stop()
	}

	c.connTimer = time.AfterFunc(1*time.Minute, func() {
		if c.IsOnline() || !c.TimeoutActive() {
			return
		}

		if onTimeout != nil {
			_ = onTimeout(context.Background(), c.ClientID())
		}
	})
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
