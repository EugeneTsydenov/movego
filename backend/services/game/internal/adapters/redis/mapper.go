package redis

import (
	"fmt"
	"game/internal/domain"
	"time"

	"github.com/google/uuid"
)

type gameDTO struct {
	ID                 string   `json:"id"`
	WhitePlayerID      string   `json:"white_player_id"`
	WhitePlayerName    string   `json:"white_player_name"`
	BlackPlayerID      string   `json:"black_player_id"`
	BlackPlayerName    string   `json:"black_player_name"`
	Status             string   `json:"status"`
	WhiteTimeRemaining int64    `json:"white_time_remaining_ns"`
	BlackTimeRemaining int64    `json:"black_time_remaining_ns"`
	TimeControlID      string   `json:"time_control_id"`
	FEN                string   `json:"fen"`
	Moves              []string `json:"moves"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
	FinishedAt         string   `json:"finished_at,omitempty"`
}

func toDomainDTO(g *domain.Game) *gameDTO {
	var moveStrs []string
	for _, m := range g.Moves() {
		moveStrs = append(moveStrs, m.String())
	}

	return &gameDTO{
		ID:                 g.ID().String(),
		WhitePlayerID:      g.WhitePlayer().ID().String(),
		WhitePlayerName:    g.WhitePlayer().Name(),
		BlackPlayerID:      g.BlackPlayer().ID().String(),
		BlackPlayerName:    g.BlackPlayer().Name(),
		Status:             string(g.Status()),
		WhiteTimeRemaining: g.WhiteTimeRemaining().Nanoseconds(),
		BlackTimeRemaining: g.BlackTimeRemaining().Nanoseconds(),
		TimeControlID:      string(g.TimeControl().ID()),
		FEN:                g.FEN(),
		Moves:              moveStrs,
		CreatedAt:          g.CreatedAt().Format(time.RFC3339),
		UpdatedAt:          g.UpdatedAt().Format(time.RFC3339),
		FinishedAt:         formatTime(g.FinishedAt()),
	}
}

func toDomainGame(dto gameDTO) (*domain.Game, error) {
	gameID, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("map game id: %v", err)
	}

	whiteID, err := uuid.Parse(dto.WhitePlayerID)
	if err != nil {
		return nil, fmt.Errorf("map white id: %v", err)
	}
	whitePlayer := domain.NewPlayer(whiteID, dto.WhitePlayerName)

	blackID, err := uuid.Parse(dto.BlackPlayerID)
	if err != nil {
		return nil, fmt.Errorf("map black id: %v", err)
	}
	blackPlayer := domain.NewPlayer(blackID, dto.BlackPlayerName)

	timeControl, err := domain.NewTimeControl(dto.TimeControlID)
	if err != nil {
		return nil, fmt.Errorf("map time control id: %v", err)
	}

	createdAt, _ := time.Parse(time.RFC3339, dto.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, dto.UpdatedAt)
	var finishedAt time.Time
	if dto.FinishedAt != "" {
		finishedAt, _ = time.Parse(time.RFC3339, dto.FinishedAt)
	}

	return domain.RestoreGame(
		gameID,
		whitePlayer,
		blackPlayer,
		domain.GameStatus(dto.Status),
		time.Duration(dto.WhiteTimeRemaining),
		time.Duration(dto.BlackTimeRemaining),
		timeControl,
		dto.FEN,
		dto.Moves,
		createdAt,
		updatedAt,
		finishedAt,
	)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
