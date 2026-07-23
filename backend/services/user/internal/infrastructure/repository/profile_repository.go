package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/movego/services/user/internal/domain/profile"
)

type ProfileRepository struct {
	db map[uuid.UUID]*profile.Profile
	mu sync.RWMutex
}

func NewProfileRepository() *ProfileRepository {
	return &ProfileRepository{
		db: make(map[uuid.UUID]*profile.Profile),
	}
}

func (r *ProfileRepository) Save(ctx context.Context, p *profile.Profile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.db[p.AccountID()] = p
	return nil
}

func (r *ProfileRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID) (*profile.Profile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.db[accountID]
	if !ok {
		return nil, profile.ErrProfileNotFound
	}
	return p, nil
}

func (r *ProfileRepository) FindByTag(ctx context.Context, tag profile.Tag) (*profile.Profile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.db {
		if p.Tag() == tag {
			return p, nil
		}
	}
	return nil, profile.ErrProfileNotFound
}

func (r *ProfileRepository) ExistsByTag(ctx context.Context, tag profile.Tag) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.db {
		if p.Tag() == tag {
			return true, nil
		}
	}
	return false, profile.ErrProfileNotFound
}
