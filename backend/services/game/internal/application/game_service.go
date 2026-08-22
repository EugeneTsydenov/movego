package application

import (
	"context"
	"game/internal/domain"
)

type GameService struct {
	gameRepository domain.GameRepository
}

func NewGameService(gameRepository domain.GameRepository) *GameService {
	return &GameService{
		gameRepository: gameRepository,
	}
}

func (s *GameService) CreateGame(ctx context.Context, in CreateGameInput) (CreateGameOutput, error) {
	whitePlayer := domain.NewPlayer(in.WhitePlayer.ID, in.WhitePlayer.Name)
	blackPlayer := domain.NewPlayer(in.BlackPlayer.ID, in.BlackPlayer.Name)
	timeControl, err := domain.NewTimeControl(in.TimeControlID)
	if err != nil {
		return CreateGameOutput{}, err
	}
	game := domain.NewGame(whitePlayer, blackPlayer, timeControl)
	err = s.gameRepository.Save(ctx, game)
	if err != nil {
		return CreateGameOutput{}, err
	}
	return CreateGameOutput{
		GameID: game.ID(),
	}, nil
}
