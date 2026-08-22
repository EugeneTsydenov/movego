package grpc

import (
	"context"
	"game/internal/application"
	gamev1 "protogen/game/v1"
)

type gameService interface {
	CreateGame(ctx context.Context, in application.CreateGameInput) (application.CreateGameOutput, error)
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

func (h *GameHandler) CreateGame(ctx context.Context, in *gamev1.CreateGameRequest) (*gamev1.CreateGameResponse, error) {
	out, err := h.gameService.CreateGame(ctx, toCreateGameInput(in))
	if err != nil {
		return nil, err
	}
	return toCreateGameResponse(out), nil
}
