package domain

import "github.com/google/uuid"

type UserID uuid.UUID

func NewUserID(idStr string) (UserID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return UserID{}, ErrInvalidUserID
	}
	return UserID(id), nil
}

func (p UserID) String() string {
	return uuid.UUID(p).String()
}

func (p UserID) UUID() uuid.UUID {
	return uuid.UUID(p)
}
