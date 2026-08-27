package game

import (
	"context"
	gamev1 "gen/game/v1"
	"matchmaking/internal/domain"
)

type Client struct {
	client gamev1.GameServiceClient
}

func NewClient(c gamev1.GameServiceClient) *Client {
	return &Client{
		client: c,
	}
}

func (c *Client) CreateGame(ctx context.Context, whitePlayer, blackPlayer *domain.Player, timeControlID domain.TimeControlID) error {
	_, err := c.client.CreateGame(ctx, toCreateGameRequest(whitePlayer, blackPlayer, timeControlID))
	if err != nil {
		return err
	}
	return nil
}
