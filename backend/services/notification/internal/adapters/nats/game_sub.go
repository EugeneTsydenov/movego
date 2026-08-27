package nats

import (
	"context"
	"log/slog"
	"shared/otelnats"

	"github.com/nats-io/nats.go/jetstream"
)

type GameJetStreamSub struct {
	js      jetstream.JetStream
	handler otelnats.MessageHanler
	logger  *slog.Logger
	cancel  context.CancelFunc
}

func NewGameJetStreamSub(js jetstream.JetStream, handler otelnats.MessageHanler, logger *slog.Logger) *GameJetStreamSub {
	return &GameJetStreamSub{
		js:      js,
		handler: handler,
		logger:  logger,
	}
}

func (s *GameJetStreamSub) Start(ctx context.Context) error {
	stream, err := s.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "GAMES",
		Subjects: []string{"game.created"},
	})
	if err != nil {
		return err
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:   "notification-service-worker",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return err
	}

	subCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	consCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		if err := s.handler(subCtx, msg); err != nil {
			s.logger.ErrorContext(subCtx, "failed to process event", "subject", msg.Subject(), "err", err)
			_ = msg.Nak()
			return
		}

		if err := msg.Ack(); err != nil {
			s.logger.ErrorContext(subCtx, "failed to jetsream message", "subject", msg.Subject(), "err", err)
		}
	})
	if err != nil {
		cancel()
		return err
	}

	go func() {
		<-subCtx.Done()
		consCtx.Stop()
	}()

	return nil
}

func (s *GameJetStreamSub) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}
