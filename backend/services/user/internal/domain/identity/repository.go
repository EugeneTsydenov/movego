package identity

import (
	"context"

	"github.com/google/uuid"
)

type CredentialRepository interface {
	Save(ctx context.Context, c *Credential) error
	FindByID(ctx context.Context, id uuid.UUID) (*Credential, error)
	FindByAccountID(ctx context.Context, accountID uuid.UUID) (*Credential, error)
	FindByEmail(ctx context.Context, email Email) (*Credential, error)
	ExistsByEmail(ctx context.Context, email Email) (bool, error)
}

type SessionRepository interface {
	Save(ctx context.Context, s *Session) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	FindAllByAccountID(ctx context.Context, accountID uuid.UUID) ([]*Session, error)
	RevokeByID(ctx context.Context, id uuid.UUID) error
	RevokeAllByAccountID(ctx context.Context, accountID uuid.UUID) error
}
