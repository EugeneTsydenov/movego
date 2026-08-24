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

func (c *Client) CreateGame(ctx context.Context, whitePlayer, blackPlayer *domain.Player, timeControlID domain.TimeControlID) (domain.GameID, error) {
	res, err := c.client.CreateGame(ctx, toCreateGameRequest(whitePlayer, blackPlayer, timeControlID))
	if err != nil {
		return domain.GameID{}, err
	}

	gameID, err := domain.NewGameID(res.GameId)
	if err != nil {
		return domain.GameID{}, err
	}

	return gameID, nil
}
