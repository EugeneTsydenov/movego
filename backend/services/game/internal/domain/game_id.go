package domain

import "github.com/google/uuid"

type GameID uuid.UUID

func NewGameID(idStr string) (GameID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return GameID{}, ErrInvalidGameID
	}
	return GameID(id), nil
}

func (p GameID) String() string {
	return uuid.UUID(p).String()
}

func (p GameID) UUID() uuid.UUID {
	return uuid.UUID(p)
}
