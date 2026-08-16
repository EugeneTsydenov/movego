package grpc

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	userIDKey contextKey = "user_id"
)

func ContextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func userIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	val := ctx.Value(userIDKey)
	if val == nil {
		return uuid.Nil, false
	}

	userID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}

	return userID, true
}
