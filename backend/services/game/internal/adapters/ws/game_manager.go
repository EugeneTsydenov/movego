package ws

import (
	"context"
	"fmt"
	"shared/wsclient"
	"sync"
)

type GameManager struct {
	mu      sync.RWMutex
	clients map[string]map[string]*wsclient.Client
}

func NewGameManager() *GameManager {
	return &GameManager{
		clients: make(map[string]map[string]*wsclient.Client),
	}
}

func (m *GameManager) Register(ctx context.Context, gameID, userID string, client *wsclient.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[gameID]; !exists {
		m.clients[gameID] = make(map[string]*wsclient.Client)
	}

	if oldClient, exists := m.clients[gameID][userID]; exists {
		oldClient.Close()
	}

	m.clients[gameID][userID] = client

	go client.WritePump(ctx)
}

func (m *GameManager) Unregister(gameID, userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if gameClients, ok := m.clients[gameID]; ok {
		if client, ok := gameClients[userID]; ok {
			client.Close()
			delete(gameClients, userID)
		}
		if len(gameClients) == 0 {
			delete(m.clients, gameID)
		}
	}
}

func (m *GameManager) SendMessage(ctx context.Context, gameID, userID string, message []byte) error {
	m.mu.RLock()
	var client *wsclient.Client
	var ok bool

	if gameClients, exists := m.clients[gameID]; exists {
		client, ok = gameClients[userID]
	}
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("client is not connected")
	}

	return client.Send(ctx, message)
}
