package memory

import (
	"context"
	"movego/internal/domain"
	"sync"

	"github.com/google/uuid"
)

type UserRepo struct {
	mu    sync.RWMutex
	users map[uuid.UUID]*domain.User
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (r *UserRepo) Save(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if u.ID() == user.ID() {
			continue
		}

		if u.Email() == user.Email() {
			return domain.ErrEmailTaken
		}

		if u.Tag() == user.Tag() {
			return domain.ErrTagTaken
		}
	}
	r.users[user.ID()] = user
	return nil
}

func (r *UserRepo) FindForAuth(
	_ context.Context,
	email domain.Email,
	_ domain.Provider,
) (*domain.User, *domain.Credential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Email() == email {
			return nil, nil, domain.ErrAuthIdentityNotFound
		}
	}

	return nil, nil, domain.ErrAuthIdentityNotFound
}
