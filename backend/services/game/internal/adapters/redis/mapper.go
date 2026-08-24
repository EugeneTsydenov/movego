package redis

import (
	"fmt"
	"game/internal/domain"
	"time"
)

type gameDTO struct {
	ID                 string   `json:"id"`
	WhitePlayerID      string   `json:"white_player_id"`
	WhitePlayerName    string   `json:"white_player_name"`
	WhitePlayerRating  int      `json:"white_player_rating"`
	BlackPlayerID      string   `json:"black_player_id"`
	BlackPlayerName    string   `json:"black_player_name"`
	BlackPlayerRating  int      `json:"black_player_rating"`
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

func toGameDTO(g *domain.Game) *gameDTO {
	var moveStrs []string
	for _, m := range g.Moves() {
		moveStrs = append(moveStrs, m.String())
	}

	return &gameDTO{
		ID:                 g.ID().String(),
		WhitePlayerID:      g.WhitePlayer().ID().String(),
		WhitePlayerName:    g.WhitePlayer().Name(),
		WhitePlayerRating:  g.WhitePlayer().Rating().Int(),
		BlackPlayerID:      g.BlackPlayer().ID().String(),
		BlackPlayerName:    g.BlackPlayer().Name(),
		BlackPlayerRating:  g.BlackPlayer().Rating().Int(),
		Status:             string(g.Status()),
		WhiteTimeRemaining: g.WhiteTimeRemaining().Nanoseconds(),
		BlackTimeRemaining: g.BlackTimeRemaining().Nanoseconds(),
		TimeControlID:      g.TimeControl().ID().String(),
		FEN:                g.FEN(),
		Moves:              moveStrs,
		CreatedAt:          g.CreatedAt().Format(time.RFC3339),
		UpdatedAt:          g.UpdatedAt().Format(time.RFC3339),
		FinishedAt:         formatTime(g.FinishedAt()),
	}
}

func toDomainGame(dto gameDTO) (*domain.Game, error) {
	gameID, err := domain.NewGameID(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("map game id: %v", err)
	}

	whiteID, err := domain.NewPlayerID(dto.WhitePlayerID)
	if err != nil {
		return nil, fmt.Errorf("map white id: %v", err)
	}

	whiteRating, err := domain.NewRating(dto.WhitePlayerRating)
	if err != nil {
		return nil, fmt.Errorf("map white rating: %v", err)
	}

	whitePlayer := domain.NewPlayer(whiteID, dto.WhitePlayerName, whiteRating)

	blackID, err := domain.NewPlayerID(dto.BlackPlayerID)
	if err != nil {
		return nil, fmt.Errorf("map black id: %v", err)
	}

	blackRating, err := domain.NewRating(dto.BlackPlayerRating)
	if err != nil {
		return nil, fmt.Errorf("map black rating: %v", err)
	}

	blackPlayer := domain.NewPlayer(blackID, dto.BlackPlayerName, blackRating)

	timeControlID, err := domain.NewTimeControlID(dto.TimeControlID)
	if err != nil {
		return nil, fmt.Errorf("map time control id: %v", err)
	}

	timeControl := domain.NewTimeControl(timeControlID)

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
