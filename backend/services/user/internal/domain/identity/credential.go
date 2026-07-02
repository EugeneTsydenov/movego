package identity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidHashFormat = errors.New("identity: invalid password hash format")

type Credential struct {
	id           uuid.UUID
	accountID    uuid.UUID
	email        Email
	passwordHash string
	createdAt    time.Time
	updatedAt    time.Time
}

func NewCredential(accountID uuid.UUID, emailStr, passwordHash string) (*Credential, error) {
	email, err := newEmail(emailStr)
	if err != nil {
		return nil, err
	}

	trimmedHash := strings.TrimSpace(passwordHash)
	if trimmedHash == "" {
		return nil, ErrInvalidHashFormat
	}

	now := time.Now().UTC()
	return &Credential{
		id:           uuid.Must(uuid.NewV7()),
		accountID:    accountID,
		email:        email,
		passwordHash: trimmedHash,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func (c *Credential) ID() uuid.UUID        { return c.id }
func (c *Credential) AccountID() uuid.UUID { return c.accountID }
func (c *Credential) Email() string        { return string(c.email) }
func (c *Credential) PasswordHash() string { return c.passwordHash }
func (c *Credential) CreatedAt() time.Time { return c.createdAt }
func (c *Credential) UpdatedAt() time.Time { return c.updatedAt }

func (c *Credential) ComparePassword(plainPassword string, compareFn func(hashed, plain string) error) bool {
	err := compareFn(c.passwordHash, plainPassword)
	return err == nil
}

func (c *Credential) ChangePassword(newPasswordHash string) error {
	trimmedHash := strings.TrimSpace(newPasswordHash)
	if trimmedHash == "" {
		return ErrInvalidHashFormat
	}

	c.passwordHash = trimmedHash
	c.updatedAt = time.Now().UTC()
	return nil
}
