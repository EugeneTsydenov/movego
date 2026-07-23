package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/movego/services/user/internal/domain/identity"
)

type SessionRepository struct {
	db map[uuid.UUID]*identity.Session
	mu sync.RWMutex
}

func NewSessionRepostiroy() *SessionRepository {
	return &SessionRepository{
		db: make(map[uuid.UUID]*identity.Session),
	}
}

func (r *SessionRepository) Save(ctx context.Context, s *identity.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.db[s.ID()] = s
	return nil
}

func (r *SessionRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*identity.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.db {
		if s.TokenHash() == tokenHash {
			return s, nil
		}
	}
	return nil, identity.ErrSessionNotFound
}

func (r *SessionRepository) FindAllByAccountID(ctx context.Context, accountID uuid.UUID) ([]*identity.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sessions := []*identity.Session{}
	for _, s := range r.db {
		if s.AccountID() == accountID {
			sessions = append(sessions, s)
		}
	}
	return sessions, nil
}
