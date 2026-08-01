package identity

import (
	"time"

	"github.com/google/uuid"
)

type Credential struct {
	id           uuid.UUID
	accountID    uuid.UUID
	email        Email
	passwordHash string
	createdAt    time.Time
	updatedAt    time.Time
}

func NewCredential(accountID uuid.UUID, email Email, hashPassword string) *Credential {
	now := time.Now().UTC()
	return &Credential{
		id:           uuid.Must(uuid.NewV7()),
		accountID:    accountID,
		email:        email,
		passwordHash: hashPassword,
		createdAt:    now,
		updatedAt:    now,
	}
}

func (c *Credential) ID() uuid.UUID        { return c.id }
func (c *Credential) AccountID() uuid.UUID { return c.accountID }
func (c *Credential) Email() Email         { return c.email }
func (c *Credential) PasswordHash() string { return c.passwordHash }
func (c *Credential) CreatedAt() time.Time { return c.createdAt }
func (c *Credential) UpdatedAt() time.Time { return c.updatedAt }

// func (c *Credential) VerifyPassword(rawPassword string) error {
// 	if bcrypt.CompareHashAndPassword([]byte(c.passwordHash), []byte(rawPassword)) != nil {
// 		return ErrInvalidCredentials
// 	}
// 	return nil
// }

// func (c *Credential) ChangePassword(newRawPassword string) error {
// 	if len(newRawPassword) < 8 {
// 		return ErrWeakPassword
// 	}
// 	hash, err := bcrypt.GenerateFromPassword([]byte(newRawPassword), bcrypt.DefaultCost)
// 	if err != nil {
// 		return err
// 	}
// 	c.passwordHash = string(hash)
// 	c.updatedAt = time.Now().UTC()
// 	return nil
// }
