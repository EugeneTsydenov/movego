package application

import (
	"context"

	"github.com/coder/websocket"
)

type Manager interface {
	Register(ctx context.Context, userID string, conn *websocket.Conn) *Client
	Unregister(userID string)
	SendMessage(ctx context.Context, userID string, message []byte) error
}
