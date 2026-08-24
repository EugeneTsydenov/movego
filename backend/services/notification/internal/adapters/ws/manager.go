package ws

import (
	"context"
	"fmt"
	"sync"
)

type Manager struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*Client),
	}
}

func (m *Manager) Register(ctx context.Context, userID string, client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if oldClient, exists := m.clients[userID]; exists {
		oldClient.Close()
	}

	m.clients[userID] = client

	go client.WritePump(ctx)
}

func (m *Manager) Unregister(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, ok := m.clients[userID]; ok {
		client.Close()
		delete(m.clients, userID)
	}
}

func (m *Manager) SendMessage(ctx context.Context, userID string, message []byte) error {
	m.mu.RLock()
	client, ok := m.clients[userID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("client is not connected")
	}

	return client.Send(ctx, message)
}
