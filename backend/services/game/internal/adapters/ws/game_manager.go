package ws

import (
	"context"
	"fmt"
	"shared/wsclient"
	"sync"
	"time"
)

type playerConnStatus struct {
	IsOnline         bool
	TimeoutActive    bool
	TimeoutExpiresAt *time.Time
}

type GameManager struct {
	mu    sync.RWMutex
	rooms map[string]*room
}

func NewGameManager() *GameManager {
	return &GameManager{
		rooms: make(map[string]*room),
	}
}

func (m *GameManager) IsRoomCreated(roomID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.rooms[roomID]
	return ok
}

// create room with fill offline clients
func (m *GameManager) CreateRoom(
	roomID string,
	clientIDs []string,
	onDisconn func(ctx context.Context, client *client) error,
	onTimeout func(ctx context.Context) error,
) {
	m.mu.Lock()
	if _, ok := m.rooms[roomID]; ok {
		m.mu.Unlock()
		return
	}

	room := newRoom(roomID)
	m.rooms[roomID] = room
	m.mu.Unlock()

	room.AddOfflineClients(clientIDs)
	room.DisconnectAll(onDisconn, onTimeout)
}

func (m *GameManager) OnPlayerConnect(
	ctx context.Context,
	roomID,
	clientID string,
	wsClient *wsclient.Client,
	onConn func(ctx context.Context, clientID string) error,
	onStart func(ctx context.Context) error,
) error {
	m.mu.RLock()
	room, ok := m.rooms[roomID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("room is not exists")
	}
	room.ConnectClient(clientID, wsClient, onConn)
	if room.IsAllConnected() {
		err := onStart(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *GameManager) DisconnectClient(
	roomID,
	clientID string,
	onDisconn func(ctx context.Context, client *client) error,
	onTimeout func(ctx context.Context) error,
) {
	m.mu.RLock()
	room, ok := m.rooms[roomID]
	m.mu.RUnlock()

	if !ok {
		return
	}
	room.DisconnectClient(clientID, onDisconn, onTimeout)
}

func (m *GameManager) SendToClient(ctx context.Context, roomID, clientID string, msg []byte) error {
	m.mu.RLock()
	room, ok := m.rooms[roomID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("room is not exists")
	}
	return room.SendToClient(ctx, clientID, msg)
}

func (m *GameManager) BroadcastToClients(ctx context.Context, roomID string, msg []byte) error {
	m.mu.RLock()
	room, ok := m.rooms[roomID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("room is not exists")
	}
	return room.BroadcastToClients(ctx, msg)
}

func (m *GameManager) GetClient(roomID, clientID string) (*client, error) {
	m.mu.RLock()
	room, ok := m.rooms[roomID]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("room is not exists")
	}

	return room.Client(clientID)
}

func (m *GameManager) IsConnected(roomID, clientID string) bool {
	m.mu.RLock()
	room, ok := m.rooms[roomID]
	m.mu.RUnlock()

	if !ok {
		return false
	}

	return room.IsConnected(clientID)
}
