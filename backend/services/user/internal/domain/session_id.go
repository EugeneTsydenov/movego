package domain

import "github.com/google/uuid"

type SessionID uuid.UUID

func NewSessionID(idStr string) (SessionID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return SessionID{}, ErrInvalidSessionID
	}
	return SessionID(id), nil
}

func (s SessionID) String() string {
	return uuid.UUID(s).String()
}

func (s SessionID) UUID() uuid.UUID {
	return uuid.UUID(s)
}
