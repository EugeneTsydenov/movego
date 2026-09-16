package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"game/internal/domain"
	"log/slog"
	"shared/otelnats"

	"github.com/nats-io/nats.go"
)

type GameCreatedEvent struct {
	GameID     string   `json:"game_id"`
	PlayersIDs []string `json:"players_ids"`
}

type GamePublisher struct {
	publishFunc otelnats.PublishFunc
	logger      *slog.Logger
}

func NewGamePublisher(publishFunc otelnats.PublishFunc, logger *slog.Logger) *GamePublisher {
	return &GamePublisher{
		publishFunc: publishFunc,
		logger:      logger,
	}
}

func (p *GamePublisher) PublishGameCreated(
	ctx context.Context,
	gameID domain.GameID,
	playersIDs []domain.PlayerID,
) error {
	event := GameCreatedEvent{
		GameID:     gameID.String(),
		PlayersIDs: make([]string, 0, len(playersIDs)),
	}

	for _, playerID := range playersIDs {
		event.PlayersIDs = append(event.PlayersIDs, playerID.String())
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err = p.publishFunc(ctx, &nats.Msg{Subject: "game.created", Data: data})
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}
