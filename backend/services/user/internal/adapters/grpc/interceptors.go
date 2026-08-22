package grpc

import (
	"context"

	"github.com/google/uuid"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ErrorInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}

		if _, ok := status.FromError(err); ok {
			return nil, err
		}

		return nil, mapDomainErrorToGRPC(err)
	}
}

func AuthContextInterceptor() grpc.UnaryServerInterceptor {
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
