package application

import (
	"context"

	"github.com/movego/services/user/internal/domain/account"
	"github.com/movego/services/user/internal/domain/authorization"
	"github.com/movego/services/user/internal/domain/identity"
	"github.com/movego/services/user/internal/domain/profile"
)

type Repositories struct {
	Accounts       account.Repository
	Credentials    identity.CredentialRepository
	Sessions       identity.SessionRepository
	Profiles       profile.Repository
	Authorizations authorization.Repository
}

type UOW interface {
	Do(ctx context.Context, fn func(*Repositories) error) error
}
