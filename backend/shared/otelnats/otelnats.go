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

// TraceExtractMiddleware extracts the trace context from the message headers and starts a span.
func TraceExtractMiddleware(tracerName string, logger *slog.Logger, next MessageHanler) MessageHanler {
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

type PublishFunc func(ctx context.Context, msg *nats.Msg, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error)

// TraceInjectMiddleware injects trace context into NATS message headers
func TraceInjectMiddleware(tracerName string, next PublishFunc) PublishFunc {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(ctx context.Context, msg *nats.Msg, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error) {
		spanCtx, span := tracer.Start(ctx, "nats.publish."+msg.Subject,
			trace.WithSpanKind(trace.SpanKindProducer),
			trace.WithAttributes(
				attribute.String("messaging.system", "nats"),
				attribute.String("messaging.destination.name", msg.Subject),
			),
		)
		defer span.End()

		if msg.Header == nil {
			msg.Header = make(nats.Header)
		}
		propagator.Inject(spanCtx, propagation.HeaderCarrier(msg.Header))

		ack, err := next(spanCtx, msg, opts...)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return ack, err
	}
}
