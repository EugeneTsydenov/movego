package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/movego/services/user/internal/domain/authorization"
)

type AuthorizationRepository struct {
	db map[uuid.UUID]*authorization.Authorization
	mu sync.RWMutex
}

func NewAuthorizationRepository() *AuthorizationRepository {
	return &AuthorizationRepository{
		db: make(map[uuid.UUID]*authorization.Authorization),
	}
}

func (r *AuthorizationRepository) Save(ctx context.Context, auth *authorization.Authorization) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.db[auth.AccountID()] = auth
	return nil
}

func (r *AuthorizationRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID) (*authorization.Authorization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	authz, ok := r.db[accountID]
	if !ok {
		return nil, authorization.ErrAuthorizationNotFound
	}
	return authz, nil
}

func (r *AuthorizationRepository) CountByRole(ctx context.Context, role authorization.Role) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cnt := 0
	for _, authz := range r.db {
		if authz.Role() == role {
			cnt++
		}
	}
	return cnt, nil
}
