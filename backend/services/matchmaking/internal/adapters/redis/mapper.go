package redis

import (
	"fmt"
	"matchmaking/internal/domain"
	"time"
)

type playerDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	TimeControlID string `json:"time_control_id"`
	Rating        int    `json:"rating"`
	JoinedAt      string `json:"joined_at"`
}

func toPlayerDTO(player *domain.Player) playerDTO {
	return playerDTO{
		ID:            player.ID().String(),
		Name:          player.Name(),
		TimeControlID: player.TimeControlID().String(),
		Rating:        player.Rating().Int(),
		JoinedAt:      player.JoinedAt().Format(time.RFC3339),
	}
}

func toDomainPlayer(dto playerDTO) (*domain.Player, error) {
	id, err := domain.NewPlayerID(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("map player id: %v", err)
	}

	timeControlID, err := domain.NewTimeControlID(dto.TimeControlID)
	if err != nil {
		return nil, fmt.Errorf("map time control id: %v", err)
	}

	rating, err := domain.NewRating(dto.Rating)
	if err != nil {
		return nil, fmt.Errorf("map rating: %v", err)
	}

	joinedAt, err := time.Parse(time.RFC3339, dto.JoinedAt)
	if err != nil {
		return nil, fmt.Errorf("map joined at: %v", err)
	}

	return domain.RestorePlayer(
		id,
		dto.Name,
		timeControlID,
		rating,
		joinedAt,
	), nil
}
