package memory

import (
	"context"
	"movego/internal/domain"
	"sync"
)

type UnitOfWork struct {
	mu sync.Mutex
}

func NewUnitOfWork() *UnitOfWork {
	return &UnitOfWork{}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(repos domain.Repos) error) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	repos := &repositories{
		users:       NewUserRepo(),
		credentials: NewCredentialRepo(),
		sessions:    NewSessionRepo(),
	}

	if err := fn(repos); err != nil {
		return err
	}

	return nil
}

type repositories struct {
	users       *UserRepo
	credentials *CredentialRepo
	sessions    *SessionRepo
}

func (r *repositories) Users() domain.UserRepo {
	return r.users
}

func (r *repositories) Credentials() domain.CredentialRepo {
	return r.credentials
}

func (r *repositories) Sessions() domain.SessionRepo {
	return r.sessions
}
