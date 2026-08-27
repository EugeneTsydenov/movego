package grpc

import (
	"context"

	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
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
