package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"notification/internal/application"

	"github.com/nats-io/nats.go/jetstream"
)

type WsSender interface {
	SendMessage(ctx context.Context, userID string, message []byte) error
}

type GameHandler struct {
	logger   *slog.Logger
	wsSender WsSender
}

func NewGameHandler(wsSender WsSender, logger *slog.Logger) *GameHandler {
	return &GameHandler{
		wsSender: wsSender,
		logger:   logger,
	}
}

func (h *GameHandler) Route(ctx context.Context, msg jetstream.Msg) error {
	switch msg.Subject() {
	case "game.created":
		return h.handleGameCreated(ctx, msg.Data())
	default:
		_ = msg.Term()
		return fmt.Errorf("unknown subject: %s", msg.Subject())
	}
}

func (h *GameHandler) handleGameCreated(ctx context.Context, data []byte) error {
	var event application.GameCreatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{
		"type":    "game_created",
		"game_id": event.GameID,
	})

	for _, id := range event.PlayersIDs {
		if err := h.wsSender.SendMessage(ctx, id, payload); err != nil {
			h.logger.Error("failed to send message", "userID", id, "error", err)
		}
	}

	return nil
}
