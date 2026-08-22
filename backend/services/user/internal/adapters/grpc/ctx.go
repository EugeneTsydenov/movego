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

func userIDFromContext(ctx context.Context) uuid.UUID {
	val := ctx.Value(userIDKey)
	userID, _ := val.(uuid.UUID)
	return userID
}
