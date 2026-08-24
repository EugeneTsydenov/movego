package application

import (
	"matchmaking/internal/domain"
)

type JoinQueueInput struct {
	PlayerID      domain.PlayerID
	PlayerName    string
	TimeControlID domain.TimeControlID
	Rating        domain.Rating
}
