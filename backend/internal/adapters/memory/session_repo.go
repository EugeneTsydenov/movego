package memory

import (
	"context"
	"movego/internal/domain"
	"sync"

	"github.com/google/uuid"
)

type SessionRepo struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*domain.Session
}

func NewSessionRepo() *SessionRepo {
	return &SessionRepo{
		sessions: make(map[uuid.UUID]*domain.Session),
	}
}

func (r *SessionRepo) Save(ctx context.Context, session *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, c := range r.sessions {
		if c.ID() == session.ID() {
			continue
		}
	}

	r.sessions[session.ID()] = session
	return nil
}
