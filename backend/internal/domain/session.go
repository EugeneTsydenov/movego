package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	id         uuid.UUID // token id
	userID     uuid.UUID
	secretHash string
	userAgent  string
	clientIP   string
	expiresAt  time.Time
	createdAt  time.Time
}

func NewSession(userID uuid.UUID, secretHash string, userAgent, clientIP string, duration time.Duration) *Session {
	now := time.Now().UTC()
	return &Session{
		id:         uuid.Must(uuid.NewV7()),
		userID:     userID,
		secretHash: secretHash,
		userAgent:  userAgent,
		clientIP:   clientIP,
		createdAt:  now,
		expiresAt:  now.Add(duration),
	}
}

func (s *Session) ID() uuid.UUID        { return s.id }
func (s *Session) UserID() uuid.UUID    { return s.userID }
func (s *Session) SecretHash() string   { return s.secretHash }
func (s *Session) UserAgent() string    { return s.userAgent }
func (s *Session) ClientIP() string     { return s.clientIP }
func (s *Session) ExpiresAt() time.Time { return s.expiresAt }
func (s *Session) CreatedAt() time.Time { return s.createdAt }
