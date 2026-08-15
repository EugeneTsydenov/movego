package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type Provider string

const (
	Google   Provider = "google"
	GitHub   Provider = "github"
	Password Provider = "password"
)

func NewProvider(providerStr string) (Provider, error) {
	switch Provider(providerStr) {
	case Google, GitHub, Password:
		return Provider(providerStr), nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidProvider, providerStr)
	}
}

func (p Provider) String() string {
	return string(p)
}

type Credential struct {
	id           uuid.UUID
	userID       uuid.UUID
	passwordHash *string
	provider     Provider
	providerKey  *string
}

func NewPasswordCredential(userID uuid.UUID, provider Provider, passwordHash string) *Credential {
	return &Credential{
		id:           uuid.Must(uuid.NewV7()),
		userID:       userID,
		provider:     provider,
		passwordHash: &passwordHash,
	}
}

func NewOAuthCredential(userID uuid.UUID, provider Provider, providerKey string) *Credential {
	return &Credential{
		id:          uuid.Must(uuid.NewV7()),
		userID:      userID,
		provider:    provider,
		providerKey: &providerKey,
	}
}

func RestoreCredential(
	id uuid.UUID,
	userID uuid.UUID,
	provider Provider,
	passwordHash *string,
	providerKey *string,
) *Credential {
	return &Credential{
		id:           id,
		userID:       userID,
		passwordHash: passwordHash,
		provider:     provider,
		providerKey:  providerKey,
	}
}

func (c *Credential) ID() uuid.UUID {
	return c.id
}

func (c *Credential) UserID() uuid.UUID {
	return c.userID
}

func (c *Credential) PasswordHash() *string {
	return c.passwordHash
}

func (c *Credential) Provider() Provider {
	return c.provider
}

func (c *Credential) ProviderKey() *string {
	return c.providerKey
}
