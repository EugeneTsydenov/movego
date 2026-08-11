package memory

import (
	"context"
	"movego/internal/domain"
	"sync"

	"github.com/google/uuid"
)

type CredentialRepo struct {
	mu          sync.RWMutex
	credentials map[uuid.UUID]*domain.Credential
}

func NewCredentialRepo() *CredentialRepo {
	return &CredentialRepo{
		credentials: make(map[uuid.UUID]*domain.Credential),
	}
}

func (r *CredentialRepo) Save(ctx context.Context, credential *domain.Credential) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, c := range r.credentials {
		if c.ID() == credential.ID() {
			continue
		}
	}

	r.credentials[credential.ID()] = credential
	return nil
}
