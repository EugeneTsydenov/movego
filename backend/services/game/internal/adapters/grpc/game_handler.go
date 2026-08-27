package grpc

import (
	"context"
	"game/internal/application"
	gamev1 "gen/game/v1"
)

type gameService interface {
	CreateGame(ctx context.Context, in application.CreateGameInput) error
}

type GameHandler struct {
	gamev1.UnimplementedGameServiceServer
	gameService gameService
}

func NewGameHandler(gameService gameService) *GameHandler {
	return &GameHandler{
		gameService: gameService,
	}
}

func (h *GameHandler) CreateGame(ctx context.Context, req *gamev1.CreateGameRequest) (*gamev1.CreateGameResponse, error) {
	in, err := toCreateGameInput(req)
	if err != nil {
		return nil, err
	}
	err = h.gameService.CreateGame(ctx, in)
	if err != nil {
		return nil, err
	}
	return &gamev1.CreateGameResponse{}, nil
}
