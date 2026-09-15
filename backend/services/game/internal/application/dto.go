package application

import (
	"game/internal/domain"
)

type PlayerDTO struct {
	ID     domain.PlayerID
	Name   string
	Rating domain.Rating
	Color  string
}

type CreateGameInput struct {
	Players       []PlayerDTO
	TimeControlID domain.TimeControlID
}

type MakeMoveOutput struct {
	Move        string
	Fen         string
	Turn        string
	WhiteTimeMs int
	BlackTimeMs int
	Status      domain.GameStatus
	Reason      string
}
