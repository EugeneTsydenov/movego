package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/movego/services/user/internal/domain/account"
)

type AccountRepository struct {
	db map[uuid.UUID]*account.Account
	mu sync.RWMutex
}

func NewAccountRepository() *AccountRepository {
	return &AccountRepository{
		db: make(map[uuid.UUID]*account.Account),
	}
}

func (r *AccountRepository) Save(ctx context.Context, account *account.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.db[account.ID()] = account
	return nil
}

func (r *AccountRepository) FindByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	acc, ok := r.db[id]
	if !ok {
		return nil, account.ErrAccountNotFound
	}
	return acc, nil
}
