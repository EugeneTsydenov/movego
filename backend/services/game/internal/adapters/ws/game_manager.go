package ws

import (
	"context"
	"fmt"
	"shared/wsclient"
	"sync"
)

type onStart func(ctx context.Context) error

type onClientConnectArgs struct {
	roomID    string
	clientID  string
	sessionID string
	wsClient  *wsclient.Client
	onConn    onConn
	onStart   onStart
}

type managerDisconnectClientArgs struct {
	roomID    string
	clientID  string
	sessionID string
	onDisconn onDisconn
	onTimeout onTimeout
}

type sendToClientArgs struct {
	roomID   string
	clientID string
	msg      []byte
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

func (m *GameManager) isRoomCreated(roomID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.rooms[roomID]

	return ok
}

// create room with fill offline clients.
func (m *GameManager) CreateRoom(
	roomID string,
	clientIDs []string,
	onTimeout onTimeout,
) {
	if m.isRoomCreated(roomID) {
		return
	}

	m.mu.Lock()
	if _, ok := m.rooms[roomID]; ok {
		m.mu.Unlock()

		return
	}

	room := newRoom(roomID)
	m.rooms[roomID] = room
	m.mu.Unlock()

	room.AddClients(clientIDs)
	room.TimeoutAll(onTimeout)
}

func (m *GameManager) OnClientConnect(
	ctx context.Context,
	args onClientConnectArgs,
) error {
	m.mu.RLock()
	room, ok := m.rooms[args.roomID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("room is not exists")
	}
	room.ConnectClient(connectClientArgs{
		clientID:  args.clientID,
		sessionID: args.sessionID,
		wsClient:  args.wsClient,
		onConn:    args.onConn,
	})
	if room.IsAllConnected() {
		err := args.onStart(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *GameManager) DisconnectClient(args managerDisconnectClientArgs) {
	m.mu.RLock()
	room, ok := m.rooms[args.roomID]
	m.mu.RUnlock()

	if !ok {
		return
	}
	room.DisconnectClient(disconnectClientArgs{
		clientID:  args.clientID,
		sessionID: args.sessionID,
		onDisconn: args.onDisconn,
		onTimeout: args.onTimeout,
	})
}

func (m *GameManager) SendToClient(ctx context.Context, args sendToClientArgs) error {
	m.mu.RLock()
	room, ok := m.rooms[args.roomID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("room is not exists")
	}

	return room.SendToClient(ctx, args.clientID, args.msg)
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
