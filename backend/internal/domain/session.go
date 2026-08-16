package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	id           uuid.UUID // token id
	userID       uuid.UUID
	secretHash   string
	userAgent    string
	clientIP     string
	lastActiveAt time.Time
	expiresAt    time.Time
	createdAt    time.Time
}

func NewSession(userID uuid.UUID, secretHash string, userAgent, clientIP string, duration time.Duration) *Session {
	now := time.Now().UTC()
	return &Session{
		id:           uuid.Must(uuid.NewV7()),
		userID:       userID,
		secretHash:   secretHash,
		userAgent:    userAgent,
		clientIP:     clientIP,
		lastActiveAt: now,
		createdAt:    now,
		expiresAt:    now.Add(duration),
	}
}

func RestoreSession(
	id,
	userID uuid.UUID,
	secretHash,
	userAgent,
	clientIP string,
	lastActiveAt,
	expiresAt,
	createdAt time.Time,
) *Session {
	return &Session{
		id:           id,
		userID:       userID,
		secretHash:   secretHash,
		userAgent:    userAgent,
		clientIP:     clientIP,
		lastActiveAt: lastActiveAt,
		expiresAt:    expiresAt,
		createdAt:    createdAt,
	}
}

func (s *Session) ID() uuid.UUID           { return s.id }
func (s *Session) UserID() uuid.UUID       { return s.userID }
func (s *Session) SecretHash() string      { return s.secretHash }
func (s *Session) UserAgent() string       { return s.userAgent }
func (s *Session) ClientIP() string        { return s.clientIP }
func (s *Session) LastActiveAt() time.Time { return s.lastActiveAt }
func (s *Session) ExpiresAt() time.Time    { return s.expiresAt }
func (s *Session) CreatedAt() time.Time    { return s.createdAt }

func (s *Session) CanBeRevoked(userID uuid.UUID) bool {
	return s.UserID() == userID
}
