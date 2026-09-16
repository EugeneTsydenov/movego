package application

import (
	"context"
	"game/internal/domain"
	"time"
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
	timeControl := domain.NewTimeControl(in.TimeControlID)

	players := make([]*domain.Player, len(in.Players))
	for i := 0; i < len(in.Players); i++ {
		playerDTO := in.Players[i]
		initialTime, _ := timeControl.Duration()
		players[i] = domain.NewPlayer(
			playerDTO.ID,
			playerDTO.Name,
			playerDTO.Rating,
			playerDTO.Color,
			initialTime,
		)
	}

	game := domain.NewGame(players, timeControl)
	err := s.gameRepo.Save(ctx, game)
	if err != nil {
		return err
	}

	playerIDs := make([]domain.PlayerID, len(players))
	for i := 0; i < len(players); i++ {
		playerIDs[i] = players[i].ID()
	}

	err = s.publisher.PublishGameCreated(ctx, game.ID(), playerIDs)
	if err != nil {
		return err
	}
	return nil
}

func (s *GameService) GetState(
	ctx context.Context,
	gameID domain.GameID,
	playerID domain.PlayerID,
) (*domain.Game, error) {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	return game, nil
}

func (s *GameService) StartGame(ctx context.Context, gameID domain.GameID, now time.Time) error {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return err
	}

	game.Start(now)

	if err := s.gameRepo.Save(ctx, game); err != nil {
		return err
	}

	return nil
}

func (s *GameService) MakeMove(
	ctx context.Context,
	gameID domain.GameID,
	playerID domain.PlayerID,
	move domain.Move,
) (MakeMoveOutput, error) {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return MakeMoveOutput{}, err
	}

	if err := game.EnsurePlayer(playerID); err != nil {
		return MakeMoveOutput{}, err
	}

	if err := game.MakeMove(playerID, move); err != nil {
		return MakeMoveOutput{}, err
	}

	if err := s.gameRepo.Save(ctx, game); err != nil {
		return MakeMoveOutput{}, err
	}

	return MakeMoveOutput{
		Move:   move.String(),
		Fen:    game.FEN(),
		Turn:   game.Position().Turn().String(),
		Status: game.Status(),
		Reason: game.Method().String(),
	}, nil
}
