package domain

import (
	"github.com/google/uuid"
)

type Credential struct {
	id           CredentialID
	userID       UserID
	passwordHash *string
	provider     Provider
	providerKey  *string
}

func NewPasswordCredential(userID UserID, provider Provider, passwordHash string) *Credential {
	return &Credential{
		id:           CredentialID(uuid.Must(uuid.NewV7())),
		userID:       userID,
		provider:     provider,
		passwordHash: &passwordHash,
	}
}

func NewOAuthCredential(userID UserID, provider Provider, providerKey string) *Credential {
	return &Credential{
		id:          CredentialID(uuid.Must(uuid.NewV7())),
		userID:      userID,
		provider:    provider,
		providerKey: &providerKey,
	}
}

func RestoreCredential(
	id CredentialID,
	userID UserID,
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

func (c *Credential) ID() CredentialID {
	return c.id
}

func (c *Credential) UserID() UserID {
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
