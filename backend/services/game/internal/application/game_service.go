package application

import (
	"context"
	"game/internal/domain"
)

type Publisher interface {
	PublishGameCreated(ctx context.Context, gameID domain.GameID, playersIDs []domain.PlayerID) error
}

type GameService struct {
	gameRepo  domain.GameRepo
	publisher Publisher
}

func NewGameService(gameRepo domain.GameRepo, publisher Publisher) *GameService {
	return &GameService{
		gameRepo:  gameRepo,
		publisher: publisher,
	}
}

func (s *GameService) CreateGame(ctx context.Context, in CreateGameInput) error {
	whitePlayer := domain.NewPlayer(in.WhitePlayer.ID, in.WhitePlayer.Name, in.WhitePlayer.Rating)
	blackPlayer := domain.NewPlayer(in.BlackPlayer.ID, in.BlackPlayer.Name, in.BlackPlayer.Rating)
	timeControl := domain.NewTimeControl(in.TimeControlID)

	game := domain.NewGame(whitePlayer, blackPlayer, timeControl)
	err := s.gameRepo.Save(ctx, game)
	if err != nil {
		return err
	}
	err = s.publisher.PublishGameCreated(ctx, game.ID(), []domain.PlayerID{whitePlayer.ID(), blackPlayer.ID()})
	if err != nil {
		return err
	}
	return nil
}
