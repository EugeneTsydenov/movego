package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"matchmaking/internal/domain"

	"github.com/nats-io/nats.go/jetstream"
)

type GameCreatedEvent struct {
	GameID     string   `json:"game_id"`
	PlayersIDs []string `json:"players_ids"`
}

type GamePublisher struct {
	js     jetstream.JetStream
	logger *slog.Logger
}

func NewGamePublisher(js jetstream.JetStream, logger *slog.Logger) *GamePublisher {
	return &GamePublisher{
		js:     js,
		logger: logger,
	}
}

func (p *GamePublisher) PublishGameCreated(ctx context.Context, gameID domain.GameID, playersIDs []domain.PlayerID) error {
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

	_, err = p.js.Publish(ctx, "game.created", data)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}
