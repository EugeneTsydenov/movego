package ws

import (
	"context"
	"fmt"
	"shared/wsclient"
	"sync"
)

type room struct {
	mu      sync.RWMutex
	roomID  string
	clients map[string]*client
}

func newRoom(roomID string) *room {
	return &room{
		roomID:  roomID,
		clients: make(map[string]*client),
	}
}
func (r *room) RoomID() string {
	return r.roomID
}

func (r *room) AddClients(clientIDs []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range clientIDs {
		if _, ok := r.clients[id]; ok {
			continue
		}
		r.clients[id] = newClient(id)
	}
}

func (r *room) TimeoutAll(
	onTimeout func(ctx context.Context, clientID string) error,
) {
	r.mu.Lock()
	clientsList := make([]*client, 0, len(r.clients))
	for _, c := range r.clients {
		clientsList = append(clientsList, c)
	}
	r.mu.Unlock()

	for _, c := range clientsList {
		c.StartInitialTimeout(onTimeout)
	}
}

func (r *room) Client(clientID string) (*client, error) {
	r.mu.RLock()
	client, ok := r.clients[clientID]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("client is not exists")
	}
	return client, nil
}

func (r *room) ConnectClient(
	clientID string,
	sessionID string,
	wsClient *wsclient.Client,
	onConn func(ctx context.Context, clientID string) error,
) {
	r.mu.Lock()
	client, ok := r.clients[clientID]
	if !ok {
		client = newClient(clientID)
		r.clients[clientID] = client
	}
	r.mu.Unlock()

	client.Connect(sessionID, wsClient, onConn)
}

func (r *room) IsAllConnected() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.clients) < 2 {
		return false
	}

	for _, c := range r.clients {
		if !c.IsOnline() {
			return false
		}
	}

	return true
}

func (r *room) IsConnected(clientID string) bool {
	r.mu.RLock()
	c, ok := r.clients[clientID]
	r.mu.RUnlock()
	if !ok {
		return false
	}
	return c.IsOnline()
}

func (r *room) DisconnectClient(
	clientID string,
	sessionID string,
	onDisconn func(ctx context.Context, client *client) error,
	onTimeout func(ctx context.Context, clientID string) error,
) {
	r.mu.RLock()
	client, ok := r.clients[clientID]
	r.mu.RUnlock()

	if !ok {
		return
	}
	_ = client.Disconnect(sessionID, onDisconn, onTimeout)
}

func (r *room) SendToClient(ctx context.Context, clientID string, msg []byte) error {
	r.mu.RLock()
	client, ok := r.clients[clientID]
	r.mu.RUnlock()

	if !ok {
		return nil
	}

	return client.Send(ctx, msg)
}

func (r *room) BroadcastToClients(ctx context.Context, msg []byte) error {
	r.mu.RLock()
	clientsList := make([]*client, 0, len(r.clients))
	for _, c := range r.clients {
		clientsList = append(clientsList, c)
	}
	r.mu.RUnlock()

	for _, c := range clientsList {
		// TODO:
		_ = c.Send(ctx, msg)
	}

	return nil
}
