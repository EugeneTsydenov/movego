package domain

import "github.com/google/uuid"

type PlayerID uuid.UUID

func NewPlayerID(idStr string) (PlayerID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return PlayerID{}, ErrInvalidUserID
	}
	return PlayerID(id), nil
}

func (p PlayerID) String() string {
	return uuid.UUID(p).String()
}

func (p PlayerID) UUID() uuid.UUID {
	return uuid.UUID(p)
}
