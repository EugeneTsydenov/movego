package application

import "github.com/google/uuid"

type PlayerDTO struct {
	ID   uuid.UUID
	Name string
}

type CreateGameInput struct {
	WhitePlayer   PlayerDTO
	BlackPlayer   PlayerDTO
	TimeControlID string
}

type CreateGameOutput struct {
	GameID uuid.UUID
}
