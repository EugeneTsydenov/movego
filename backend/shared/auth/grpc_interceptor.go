package auth

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ContextInterceptor is a gRPC interceptor that extracts the user ID from the incoming context.
func ContextInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get("x-user-id")
		if len(values) == 0 || values[0] == "" {
			return nil, status.Error(codes.Unauthenticated, "missing user identity")
		}

		userID, err := uuid.Parse(values[0])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid user identifier")
		}

		ctx = ContextWithUserID(ctx, userID)
		return handler(ctx, req)
	}
}
