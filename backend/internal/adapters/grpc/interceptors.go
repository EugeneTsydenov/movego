package grpc

import (
	"context"
	"errors"
	"log/slog"

	"buf.build/go/protovalidate"
	"github.com/google/uuid"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
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

func LoggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			st, _ := status.FromError(err)

			switch st.Code() {
			case codes.InvalidArgument, codes.Unauthenticated, codes.NotFound, codes.AlreadyExists, codes.PermissionDenied:
				log.InfoContext(ctx, "grpc business error",
					slog.String("method", info.FullMethod),
					slog.String("code", st.Code().String()),
					slog.String("error", err.Error()),
				)
			default:
				log.ErrorContext(ctx, "system error",
					slog.String("method", info.FullMethod),
					slog.String("code", st.Code().String()),
					slog.String("error", err.Error()),
				)
			}

			return resp, err
		}

		log.InfoContext(ctx, "grpc request success",
			slog.String("method", info.FullMethod),
		)
		return resp, err
	}
}

func ValidationUnaryInterceptor(v protovalidate.Validator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if msg, ok := req.(proto.Message); ok {
			if err := v.Validate(msg); err != nil {
				var valErr *protovalidate.ValidationError
				if errors.As(err, &valErr) {
					br := &errdetails.BadRequest{}

					for _, violation := range valErr.Violations {
						br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
							Field:       string(violation.FieldDescriptor.Name()),
							Description: violation.String(),
						})
					}

					st := status.New(codes.InvalidArgument, err.Error())
					stWithDetails, _ := st.WithDetails(br)
					return nil, stWithDetails.Err()
				}

				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}

		return handler(ctx, req)
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
