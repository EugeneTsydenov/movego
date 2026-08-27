package main

import (
	"context"
	"fmt"
	"game/internal/application"
	"game/internal/config"
	gamev1 "gen/game/v1"
	"log/slog"
	"net"
	sharedinterceptor "shared/interceptor"
	"shared/logger"
	"shared/otelnats"
	sharedredis "shared/redis"
	"shared/telemetry"

	grpcadapter "game/internal/adapters/grpc"
	natsadapter "game/internal/adapters/nats"
	redisadapter "game/internal/adapters/redis"

	"buf.build/go/protovalidate"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type app struct {
	cfg          *config.Config
	logger       *slog.Logger
	shutdownOtel func(context.Context) error
	redisClient  *redis.Client
	natsConn     *nats.Conn
	listener     net.Listener
	grpcServer   *grpc.Server
}

func newApp(ctx context.Context, cfg *config.Config, env string) (*app, error) {
	appLogger := logger.New(env, logger.FromStringLevel(env))
	shutdownOtel, err := telemetry.Init(ctx, cfg.App.Name, cfg.Otel.Endpoint, cfg.Otel.MetricsPort)
	if err != nil {
		return nil, err
	}

	redisClient, err := sharedredis.NewClient(ctx, &redis.Options{
		Addr:            cfg.Redis.Addr,
		Username:        cfg.Redis.Username,
		Password:        cfg.Redis.Password,
		DB:              cfg.Redis.DB,
		PoolSize:        cfg.Redis.PoolSize,
		MinIdleConns:    cfg.Redis.MinIdleConns,
		ConnMaxIdleTime: cfg.Redis.ConnMaxIdleTime,
		DialTimeout:     cfg.Redis.DialTimeout,
		ReadTimeout:     cfg.Redis.ReadTimeout,
		WriteTimeout:    cfg.Redis.WriteTimeout,
		MaxRetries:      cfg.Redis.MaxRetries,
	})
	if err != nil {
		return nil, err
	}

	nc, err := nats.Connect(cfg.Nats.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	validator, _ := protovalidate.New()
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			sharedinterceptor.RecoveryInterceptor(appLogger),
			sharedinterceptor.LoggingInterceptor(appLogger),
			grpcadapter.ErrorInterceptor(),
			sharedinterceptor.ValidationUnaryInterceptor(validator),
		),
	)
	reflection.Register(grpcServer)

	return &app{
		cfg:          cfg,
		logger:       appLogger,
		shutdownOtel: shutdownOtel,
		redisClient:  redisClient,
		natsConn:     nc,
		listener:     lis,
		grpcServer:   grpcServer,
	}, nil
}

func (a *app) InitDeps() error {
	gameRepo := redisadapter.NewGameRepo(a.redisClient)
	js, err := jetstream.New(a.natsConn)
	if err != nil {
		return err
	}
	basePublish := func(ctx context.Context, msg *nats.Msg, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error) {
		return js.PublishMsg(ctx, msg, opts...)
	}
	tracedPublish := otelnats.TraceInjectMiddleware(a.cfg.App.Name, basePublish)
	gamePublisher := natsadapter.NewGamePublisher(tracedPublish, a.logger)
	gameService := application.NewGameService(gameRepo, gamePublisher)
	gameHandler := grpcadapter.NewGameHandler(gameService)
	gamev1.RegisterGameServiceServer(a.grpcServer, gameHandler)
	return nil
}

func (a *app) Run() error {
	g := new(errgroup.Group)
	g.Go(func() error {
		return a.grpcServer.Serve(a.listener)
	})
	return g.Wait()
}

func (a *app) Shutdown(ctx context.Context) {
	stopped := make(chan struct{})
	go func() {
		if a.grpcServer != nil {
			a.grpcServer.GracefulStop()
		}
		close(stopped)
	}()

	select {
	case <-stopped:
		a.logger.Info("gRPC server gracefully stopped")
	case <-ctx.Done():
		a.logger.Warn("shutdown timeout")
		if a.grpcServer != nil {
			a.grpcServer.Stop()
		}
	}

	if a.shutdownOtel != nil {
		if err := a.shutdownOtel(ctx); err != nil {
			a.logger.Error("failed to shutdown otel", "error", err)
		}
	}
	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			a.logger.Error("failed to close redis client", "error", err)
		}
	}
	if a.natsConn != nil {
		a.natsConn.Close()
	}
	if a.listener != nil {
		if err := a.listener.Close(); err != nil {
			a.logger.Error("failed to close listener", "error", err)
		}
	}
}
