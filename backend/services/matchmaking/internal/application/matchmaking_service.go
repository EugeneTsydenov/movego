package application

import (
	"context"
	"matchmaking/internal/domain"
)

type MatchmakingService struct {
	queueRepo domain.QueueRepo
}

func NewMatchmakingService(queueRepo domain.QueueRepo) *MatchmakingService {
	return &MatchmakingService{
		queueRepo: queueRepo,
	}
}

func (s *MatchmakingService) JoinQueue(ctx context.Context, in JoinQueueInput) error {
	exists, err := s.queueRepo.Exists(ctx, in.PlayerID)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrAlreadyInQueue
	}

	player := domain.NewPlayer(in.PlayerID, in.PlayerName, in.TimeControlID, in.Rating)
	if err := s.queueRepo.Save(ctx, player); err != nil {
		return err
	}
	return nil
}
