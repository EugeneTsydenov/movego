package application

import (
	"context"
	"movego/internal/domain"

	"github.com/google/uuid"
)

type UnitOfWork interface {
	Do(ctx context.Context, fn func(repos domain.Repos) error) error
}

type TokenClaims struct {
	Sub         uuid.UUID
	SessionID   uuid.UUID
	Roles       []string
	Permissions []string
}

type TokenIssuer interface {
	Issue(claims TokenClaims) (string, error)
	Verify(token string) (TokenClaims, error)
}

type Hasher interface {
	Hash(val string) (string, error)
	Verify(raw, hash string) (bool, error)
}
