package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/movego/services/user/internal/application"
)

type Issuer struct {
	secret []byte
	ttl    time.Duration
}

func NewIssuer(secret []byte, ttl time.Duration) *Issuer {
	return &Issuer{secret, ttl}
}

func (i *Issuer) IssueAccessToken(ctx context.Context, claims application.TokenClaims) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(i.ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  claims.AccountID.String(),
		"role": claims.Role,
		"exp":  expiresAt.Unix(),
	})
	signed, err := token.SignedString(i.secret)
	return signed, expiresAt, err
}

func (i *Issuer) ParseAccessToken(ctx context.Context, token string) (application.TokenClaims, error)
