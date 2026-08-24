package application

import (
	"context"
	"game/internal/domain"
)

type GameService struct {
	gameRepo domain.GameRepo
}

func NewGameService(gameRepo domain.GameRepo) *GameService {
	return &GameService{
		gameRepo: gameRepo,
	}
}

func (s *GameService) CreateGame(ctx context.Context, in CreateGameInput) (CreateGameOutput, error) {
	whitePlayer := domain.NewPlayer(in.WhitePlayer.ID, in.WhitePlayer.Name, in.WhitePlayer.Rating)
	blackPlayer := domain.NewPlayer(in.BlackPlayer.ID, in.BlackPlayer.Name, in.BlackPlayer.Rating)
	timeControl := domain.NewTimeControl(in.TimeControlID)

	game := domain.NewGame(whitePlayer, blackPlayer, timeControl)
	err := s.gameRepo.Save(ctx, game)
	if err != nil {
		return CreateGameOutput{}, err
	}
	return CreateGameOutput{
		GameID: game.ID(),
	}, nil
}
