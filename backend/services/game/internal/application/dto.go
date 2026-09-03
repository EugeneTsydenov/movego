package application

import (
	"game/internal/domain"
)

type PlayerDTO struct {
	ID     domain.PlayerID
	Name   string
	Rating domain.Rating
}

type CreateGameInput struct {
	WhitePlayer   PlayerDTO
	BlackPlayer   PlayerDTO
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
