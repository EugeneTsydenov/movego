package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	id           SessionID // token id
	userID       UserID
	secretHash   string
	userAgent    string
	clientIP     string
	lastActiveAt time.Time
	expiresAt    time.Time
	createdAt    time.Time
}

func NewSession(userID UserID, secretHash string, userAgent, clientIP string, duration time.Duration) *Session {
	now := time.Now().UTC()
	return &Session{
		id:           SessionID(uuid.Must(uuid.NewV7())),
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
	id SessionID,
	userID UserID,
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

func (s *Session) ID() SessionID           { return s.id }
func (s *Session) UserID() UserID          { return s.userID }
func (s *Session) SecretHash() string      { return s.secretHash }
func (s *Session) UserAgent() string       { return s.userAgent }
func (s *Session) ClientIP() string        { return s.clientIP }
func (s *Session) LastActiveAt() time.Time { return s.lastActiveAt }
func (s *Session) ExpiresAt() time.Time    { return s.expiresAt }
func (s *Session) CreatedAt() time.Time    { return s.createdAt }

func (s *Session) CanBeRevoked(userID UserID) bool {
	return s.UserID() == userID
}
