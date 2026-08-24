package domain

import "github.com/google/uuid"

type CredentialID uuid.UUID

func NewCredentialID(idStr string) (CredentialID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return CredentialID{}, ErrInvalidCredentialID
	}
	return CredentialID(id), nil
}

func (c CredentialID) String() string {
	return uuid.UUID(c).String()
}

func (c CredentialID) UUID() uuid.UUID {
	return uuid.UUID(c)
}
