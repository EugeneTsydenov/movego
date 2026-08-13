package grpc

import (
	"context"
	"errors"
	"log/slog"
	"movego/internal/domain"

	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrorHInterceptor() grpc.UnaryServerInterceptor {
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

		switch {
		case errors.Is(err, domain.ErrNotFound):
			return nil, status.Error(grpccodes.NotFound, err.Error())
		case errors.Is(err, domain.ErrAlreadyExists):
			return nil, status.Error(grpccodes.AlreadyExists, err.Error())
		case errors.Is(err, domain.ErrValidation):
			return nil, status.Error(grpccodes.InvalidArgument, err.Error())
		}

		return nil, status.Error(grpccodes.Internal, "internal server error")
	}
}

func LoggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			log.WarnContext(ctx, "grpc request failed",
				slog.String("method", info.FullMethod),
				slog.String("error", err.Error()),
			)
		} else {
			log.InfoContext(ctx, "grpc request success",
				slog.String("method", info.FullMethod),
			)
		}
		return resp, err
	}
}
