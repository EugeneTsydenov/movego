package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserRepo interface {
	Save(ctx context.Context, user *User) error
}

type CredentialRepo interface {
	Save(ctx context.Context, cred *Credential) error
	FindForAuth(ctx context.Context, email Email, provider Provider) (*User, *Credential, error)
}

type SessionRepo interface {
	Save(ctx context.Context, session *Session) error
	FindValid(ctx context.Context, id uuid.UUID) (*Session, *User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Session, error)
	ListActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Repos interface {
	Users() UserRepo
	Credentials() CredentialRepo
	Sessions() SessionRepo
}
