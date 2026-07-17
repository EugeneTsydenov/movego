package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func (i *Issuer) ParseAccessToken(ctx context.Context, tokenStr string) (application.TokenClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("jwt: unexpected signing method")
		}
		return i.secret, nil
	})
	if err != nil {
		return application.TokenClaims{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return application.TokenClaims{}, errors.New("jwt: invalid token")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return application.TokenClaims{}, errors.New("jwt: missing sub claim")
	}
	accountID, err := uuid.Parse(sub)
	if err != nil {
		return application.TokenClaims{}, err
	}

	role, _ := claims["role"].(string)

	return application.TokenClaims{
		AccountID: accountID,
		Role:      role,
	}, nil
}
