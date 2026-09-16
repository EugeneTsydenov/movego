package main

import (
	"context"
	"fmt"
	"log/slog"
	"matchmaking/internal/adapters/game"
	"matchmaking/internal/application"
	"matchmaking/internal/config"
	"net"
	"shared/auth"
	"shared/telemetry"

	"buf.build/go/protovalidate"
	gamev1 "gen/game/v1"
	matchmakingv1 "gen/matchmaking/v1"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	grpcadapter "matchmaking/internal/adapters/grpc"
	redisadapter "matchmaking/internal/adapters/redis"

	sharedinterceptor "shared/interceptor"
	sharedredis "shared/redis"
)

type app struct {
	logger            *slog.Logger
	shutdownOtel      func(context.Context) error
	redisClient       *redis.Client
	matchmakingWorker *application.MatchmakingWorker
	gameConn          *grpc.ClientConn
	listner           net.Listener
	server            *grpc.Server
}

func newApp(ctx context.Context, logger *slog.Logger, cfg *config.Config, env string) (*app, error) {
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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	validator, _ := protovalidate.New()
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			sharedinterceptor.RecoveryInterceptor(logger),
			sharedinterceptor.LoggingInterceptor(logger),
			grpcadapter.ErrorInterceptor(),
			sharedinterceptor.ValidationUnaryInterceptor(validator),
			auth.ContextInterceptor(),
		),
	)

	gameConn, err := initGameConn(ctx, cfg)
	if err != nil {
		return nil, err
	}

	reflection.Register(grpcServer)

	return &app{
		server:       grpcServer,
		listner:      lis,
		redisClient:  redisClient,
		gameConn:     gameConn,
		shutdownOtel: shutdownOtel,
		logger:       logger,
	}, nil
}

func initGameConn(ctx context.Context, cfg *config.Config) (*grpc.ClientConn, error) {
	target := fmt.Sprintf("%s:%d", cfg.GameClient.Host, cfg.GameClient.Port)
	var opts []grpc.DialOption

	if cfg.GameClient.TlsEnabled {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	opts = append(opts, grpc.WithStatsHandler(otelgrpc.NewClientHandler()))

	conn, err := grpc.NewClient(
		target,
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}
	return conn, nil
}

func (a *app) InitDeps() {
	queueRepo := redisadapter.NewQueueRepo(a.redisClient)
	c := gamev1.NewGameServiceClient(a.gameConn)
	gameClient := game.NewClient(c)
	matchmakingService := application.NewMatchmakingService(queueRepo)
	a.matchmakingWorker = application.NewMatchmakingWorker(a.logger, queueRepo, gameClient)
	matchmakingv1.RegisterMatchmakingServiceServer(a.server, grpcadapter.NewMatchmakingHandler(matchmakingService))
}

func (a *app) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return a.server.Serve(a.listner)
	})
	g.Go(func() error {
		a.logger.Info("matchmaking worker started")
		a.matchmakingWorker.Start(ctx)
		return nil
	})
	return g.Wait()
}

func (a *app) Shutdown(ctx context.Context) {
	stopped := make(chan struct{})
	go func() {
		if a.server != nil {
			a.server.GracefulStop()
		}
		close(stopped)
	}()

	select {
	case <-stopped:
		a.logger.Info("gRPC server gracefully stopped")
	case <-ctx.Done():
		a.logger.Warn("shutdown timeout")
		if a.server != nil {
			a.server.Stop()
		}
	}

	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			a.logger.Error("failed to close redis client", "error", err)
		}
	}

	if a.shutdownOtel != nil {
		if err := a.shutdownOtel(ctx); err != nil {
			a.logger.Error("failed to shutdown otel", "error", err)
		}
	}

	if a.gameConn != nil {
		if err := a.gameConn.Close(); err != nil {
			a.logger.Error("failed to close game client", "error", err)
		}
	}
}
