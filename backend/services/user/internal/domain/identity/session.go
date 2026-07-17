package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSessionRevoked  = errors.New("identity: session revoked")
	ErrSessionExpired  = errors.New("identity: session expired")
	ErrSessionNotFound = errors.New("identity: session not found")
)

type Session struct {
	id        uuid.UUID
	accountID uuid.UUID
	tokenHash string
	userAgent string
	clientIP  string
	isRevoked bool
	expiresAt time.Time
	createdAt time.Time
	revokedAt *time.Time
}

func NewSession(accountID uuid.UUID, tokenHash, userAgent, clientIP string, duration time.Duration) *Session {
	now := time.Now().UTC()
	return &Session{
		id:        uuid.Must(uuid.NewV7()),
		accountID: accountID,
		tokenHash: tokenHash,
		userAgent: userAgent,
		clientIP:  clientIP,
		createdAt: now,
		expiresAt: now.Add(duration),
	}
}

func (s *Session) ID() uuid.UUID         { return s.id }
func (s *Session) AccountID() uuid.UUID  { return s.accountID }
func (s *Session) TokenHash() string     { return s.tokenHash }
func (s *Session) UserAgent() string     { return s.userAgent }
func (s *Session) ClientIP() string      { return s.clientIP }
func (s *Session) IsRevoked() bool       { return s.isRevoked }
func (s *Session) ExpiresAt() time.Time  { return s.expiresAt }
func (s *Session) CreatedAt() time.Time  { return s.createdAt }
func (s *Session) RevokedAt() *time.Time { return s.revokedAt }

func (s *Session) Revoke() {
	if s.isRevoked {
		return
	}

	now := time.Now().UTC()
	s.isRevoked = true
	s.revokedAt = &now
}

func (s *Session) Validate() error {
	if s.isRevoked {
		return ErrSessionRevoked
	}

	if time.Now().UTC().After(s.expiresAt) {
		return ErrSessionExpired
	}

	return nil
}

func GenerateRefreshToken() (rawToken string, tokenHash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	rawToken = base64.RawURLEncoding.EncodeToString(buf)
	tokenHash = HashRefreshToken(rawToken)
	return rawToken, tokenHash, nil
}

func HashRefreshToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}
