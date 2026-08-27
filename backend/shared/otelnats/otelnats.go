package otelnats

import (
	"context"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type MessageHanler func(ctx context.Context, msg jetstream.Msg) error

func TraceMiddleware(tracerName string, logger *slog.Logger, next MessageHanler) MessageHanler {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()
	return func(ctx context.Context, msg jetstream.Msg) error {
		defer func() {
			if r := recover(); r != nil {
				logger.ErrorContext(ctx, "recovered from panic in nats consumer",
					slog.Any("panic", r),
					slog.String("subject", msg.Subject()),
				)
			}
		}()

		msgHeaders := msg.Headers()
		if msgHeaders == nil {
			msgHeaders = make(nats.Header)
		}
		parentCtx := propagator.Extract(ctx, propagation.HeaderCarrier(msgHeaders))

		spanCtx, span := tracer.Start(parentCtx, "nats."+msg.Subject(),
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(
				attribute.String("messaging.system", "nats"),
				attribute.String("messaging.destination.name", msg.Subject()),
			),
		)
		defer span.End()

		err := next(spanCtx, msg)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return err
	}
}
