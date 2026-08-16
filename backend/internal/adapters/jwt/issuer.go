package jwt

import (
	"errors"
	"fmt"
	"time"

	"movego/internal/application"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

type customClaims struct {
	jwt.RegisteredClaims
	SessionID   string   `json:"session_id,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type Issuer struct {
	secretKey []byte
	ttl       time.Duration
	issuer    string
}

func NewIssuer(secretKey []byte, ttl time.Duration, issuer string) *Issuer {
	return &Issuer{
		secretKey: secretKey,
		ttl:       ttl,
		issuer:    issuer,
	}
}

func (i *Issuer) Issue(claims application.TokenClaims) (string, error) {
	now := time.Now().UTC()

	jwtClaims := customClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   claims.Sub.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		SessionID:   claims.SessionID.String(),
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	signedToken, err := token.SignedString(i.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (i *Issuer) Verify(tokenStr string) (application.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &customClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return i.secretKey, nil
	})

	if err != nil || !token.Valid {
		return application.TokenClaims{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok {
		return application.TokenClaims{}, ErrInvalidToken
	}

	subUUID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return application.TokenClaims{}, fmt.Errorf("invalid subject UUID: %w", err)
	}

	sessionUUID, err := uuid.Parse(claims.SessionID)
	if err != nil {
		return application.TokenClaims{}, fmt.Errorf("invalid session ID UUID: %w", err)
	}

	return application.TokenClaims{
		Sub:         subUUID,
		SessionID:   sessionUUID,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	}, nil
}
