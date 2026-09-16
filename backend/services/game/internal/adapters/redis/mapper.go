package redis

import (
	"fmt"
	"game/internal/domain"
	"time"
)

type playerDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Rating        int    `json:"rating"`
	Color         string `json:"color"`
	TimeRemaining int64  `json:"time_remaining_ns"`
}

type gameDTO struct {
	ID            string      `json:"id"`
	Players       []playerDTO `json:"players"`
	Status        string      `json:"status"`
	TimeControlID string      `json:"time_control_id"`
	FEN           string      `json:"fen"`
	Moves         []string    `json:"moves"`
	CreatedAt     string      `json:"created_at"`
	UpdatedAt     string      `json:"updated_at"`
	FinishedAt    string      `json:"finished_at,omitempty"`
}

func toPlayerDTO(player *domain.Player) playerDTO {
	return playerDTO{
		ID:            player.ID().String(),
		Name:          player.Name(),
		Rating:        player.Rating().Int(),
		Color:         player.Color(),
		TimeRemaining: player.TimeRemaining().Nanoseconds(),
	}
}

func toPlayerDTOs(players []*domain.Player) []playerDTO {
	dtos := make([]playerDTO, len(players))
	for i := 0; i < len(players); i++ {
		dtos[i] = toPlayerDTO(players[i])
	}

	return dtos
}

func toGameDTO(g *domain.Game) *gameDTO {
	var moveStrs []string
	for _, m := range g.Moves() {
		moveStrs = append(moveStrs, m.String())
	}

	return &gameDTO{
		ID:            g.ID().String(),
		Players:       toPlayerDTOs(g.Players()),
		Status:        string(g.Status()),
		TimeControlID: g.TimeControl().ID().String(),
		FEN:           g.FEN(),
		Moves:         moveStrs,
		CreatedAt:     g.CreatedAt().Format(time.RFC3339),
		UpdatedAt:     g.UpdatedAt().Format(time.RFC3339),
		FinishedAt:    formatTime(g.FinishedAt()),
	}
}

func toDomainPlayer(dto playerDTO) (*domain.Player, error) {
	id, err := domain.NewPlayerID(dto.ID)
	if err != nil {
		return nil, err
	}
	return domain.NewPlayer(
		id,
		dto.Name,
		domain.Rating(dto.Rating),
		dto.Color,
		time.Duration(dto.TimeRemaining),
	), nil
}

func toDomainPlayers(dtos []playerDTO) ([]*domain.Player, error) {
	players := make([]*domain.Player, len(dtos))
	for i := 0; i < len(dtos); i++ {
		player, err := toDomainPlayer(dtos[i])
		if err != nil {
			return nil, err
		}
		players[i] = player
	}

	return players, nil
}

func toDomainGame(dto gameDTO) (*domain.Game, error) {
	gameID, err := domain.NewGameID(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("map game id: %v", err)
	}

	players, err := toDomainPlayers(dto.Players)
	if err != nil {
		return nil, fmt.Errorf("map players: %v", err)
	}

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
		players,
		domain.GameStatus(dto.Status),
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
