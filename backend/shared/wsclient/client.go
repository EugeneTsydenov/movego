package wsclient

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type Client struct {
	conn      *websocket.Conn
	send      chan []byte
	closeOnce sync.Once
}

func New(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}
}

func (c *Client) Send(ctx context.Context, message []byte) error {
	select {
	case c.send <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return errors.New("client buffer is full, connection is slow")
	}
}

func (c *Client) WritePump(ctx context.Context) {
	defer c.conn.CloseNow()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}

			writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()

			if err != nil {
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.send)
		c.conn.CloseNow()
	})
}
