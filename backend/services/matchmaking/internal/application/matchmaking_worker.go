package application

import (
	"context"
	"log/slog"
	"matchmaking/internal/domain"
	"time"
)

type GameClient interface {
	CreateGame(ctx context.Context, whitePlayer, blackPlayer *domain.Player, timeControlID domain.TimeControlID) (domain.GameID, error)
}

type Publisher interface {
	PublishGameCreated(ctx context.Context, gameID domain.GameID, playersIDs []domain.PlayerID) error
}

type MatchmakingWorker struct {
	logger     *slog.Logger
	queueRepo  domain.QueueRepo
	gameClient GameClient
	publisher  Publisher
}

func NewMatchmakingWorker(logger *slog.Logger, queueRepo domain.QueueRepo, gameClient GameClient, publisher Publisher) *MatchmakingWorker {
	return &MatchmakingWorker{
		logger:     logger,
		queueRepo:  queueRepo,
		gameClient: gameClient,
		publisher:  publisher,
	}
}

func (w *MatchmakingWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.logger.Debug("matchmaking tick")
			w.proccessMatching(ctx)
		}
	}
}

func (w *MatchmakingWorker) proccessMatching(ctx context.Context) {
	players, err := w.queueRepo.FindAll(ctx)
	if err != nil {
		w.logger.Error("failed to fetch queue", "err", err)
		return
	}

	for i := 0; i < len(players); i++ {
		p1 := players[i]
		for j := i + 1; j < len(players); j++ {
			p2 := players[j]

			if p1.CanMatch(p2) {
				w.logger.Info("match found, trying to lock players", "p1", p1.ID(), "p2", p2.ID())

				err := w.queueRepo.Delete(ctx, p1.ID(), p2.ID())
				if err != nil {
					w.logger.Warn("failed to delete players from queue (maybe already matched)", "err", err)
					continue
				}

				white, black := domain.AssignColors(p1, p2)

				gameID, err := w.gameClient.CreateGame(ctx, white, black, white.TimeControlID())
				if err != nil {
					w.logger.Error("failed to create game, players are lost from queue!", "err", err, "p1", p1.ID(), "p2", p2.ID())
					continue
				}

				w.logger.Info("game created successfully", "gameID", gameID)

				if err = w.publisher.PublishGameCreated(ctx, gameID, []domain.PlayerID{p1.ID(), p2.ID()}); err != nil {
					w.logger.Error("failed to publish game created event", "err", err)
					continue
				}

				return
			}
		}
	}
}
