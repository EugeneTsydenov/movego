package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/movego/services/user/internal/domain/identity"
)

type CredentialRepository struct {
	db map[uuid.UUID]*identity.Credential
	mu sync.RWMutex
}

func NewCredentialRepository() *CredentialRepository {
	return &CredentialRepository{
		db: make(map[uuid.UUID]*identity.Credential),
	}
}

func (r *CredentialRepository) Save(ctx context.Context, c *identity.Credential) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.db[c.ID()] = c
	return nil
}

func (r *CredentialRepository) FindByID(ctx context.Context, id uuid.UUID) (*identity.Credential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c := r.db[id]
	if c == nil {
		return nil, identity.ErrCredentialNotFound
	}
	return c, nil
}

func (r *CredentialRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID) (*identity.Credential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.db {
		if c.AccountID() == accountID {
			return c, nil
		}
	}
	return nil, identity.ErrCredentialNotFound
}

func (r *CredentialRepository) FindByEmail(ctx context.Context, email identity.Email) (*identity.Credential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.db {
		if c.Email() == email {
			return c, nil
		}
	}
	return nil, identity.ErrCredentialNotFound
}

func (r *CredentialRepository) ExistsByEmail(ctx context.Context, email identity.Email) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.db {
		if c.Email() == email {
			return true, nil
		}
	}
	return false, nil
}
