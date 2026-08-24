package main

import (
	"context"
	"fmt"
	gamev1 "gen/game/v1"
	matchmakingv1 "gen/matchmaking/v1"
	"log/slog"
	"net"

	"matchmaking/internal/adapters/game"
	grpcadapter "matchmaking/internal/adapters/grpc"
	redisadapter "matchmaking/internal/adapters/redis"
	"matchmaking/internal/application"
	"matchmaking/internal/config"

	sharedinterceptor "shared/interceptor"
	"shared/logger"
	"shared/telemetry"

	"buf.build/go/protovalidate"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type app struct {
	server            *grpc.Server
	listner           net.Listener
	redisClient       *redis.Client
	gameConn          *grpc.ClientConn
	shutdownOtel      func(context.Context) error
	Logger            *slog.Logger
	matchmakingWorker *application.MatchmakingWorker
}

func newApp(ctx context.Context, cfg *config.Config, env string) (*app, error) {
	appLogger := initLogger(cfg, env)

	shutdownOtel, err := initTelemetry(ctx, cfg, appLogger)
	if err != nil {
		return nil, err
	}

	redisClient, err := initRedis(ctx, cfg, appLogger)
	if err != nil {
		return nil, err
	}

	lis, err := initListener(cfg.Server.Port, appLogger)
	if err != nil {
		return nil, err
	}

	server := initGRPCServer(appLogger)

	gameConn, err := initGameConn(ctx, cfg, appLogger)
	if err != nil {
		return nil, err
	}

	reflection.Register(server)

	appLogger.Info("gRPC infrastructure registered successfully")

	return &app{
		server:       server,
		listner:      lis,
		redisClient:  redisClient,
		gameConn:     gameConn,
		shutdownOtel: shutdownOtel,
		Logger:       appLogger,
	}, nil
}

func (a *app) initModules() {
	queueRepo := redisadapter.NewQueueRepo(a.redisClient)
	c := gamev1.NewGameServiceClient(a.gameConn)
	gameClient := game.NewClient(c)
	matchmakingService := application.NewMatchmakingService(queueRepo)
	a.matchmakingWorker = application.NewMatchmakingWorker(a.Logger, queueRepo, gameClient)
	matchmakingv1.RegisterMatchmakingServiceServer(a.server, grpcadapter.NewMatchmakingHandler(matchmakingService))
}

func initLogger(cfg *config.Config, env string) *slog.Logger {
	appLogger := logger.New(env, logger.FromStringLevel(cfg.App.LogLevel))
	appLogger.Info("initializing application", "app_name", cfg.App.Name, "env", env)
	return appLogger
}

func initTelemetry(ctx context.Context, cfg *config.Config, appLogger *slog.Logger) (func(context.Context) error, error) {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		appLogger.Error("opentelemetry internal error", "error", err)
	}))

	shutdownOtel, err := telemetry.InitTelemetry(ctx, cfg.App.Name, cfg.Otel.Endpoint, cfg.Otel.MetricsPort)
	if err != nil {
		appLogger.Error("failed to init telemetry", "error", err)
		return nil, fmt.Errorf("failed to init telemetry: %w", err)
	}

	appLogger.Info("telemetry initialized", "endpoint", cfg.Otel.Endpoint, "metrics_port", cfg.Otel.MetricsPort)
	return shutdownOtel, nil
}

func initRedis(ctx context.Context, cfg *config.Config, appLogger *slog.Logger) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
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

	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		return nil, fmt.Errorf("failed to instrument redis tracing: %w", err)
	}

	if err := redisotel.InstrumentMetrics(redisClient); err != nil {
		return nil, fmt.Errorf("failed to instrument redis metrics: %w", err)
	}

	if err := redisClient.Ping(ctx).Err(); err != nil {
		appLogger.Error("failed to ping redis", "addr", cfg.Redis.Addr, "error", err)
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	appLogger.Info("connected to redis successfully", "addr", cfg.Redis.Addr, "db", cfg.Redis.DB)
	return redisClient, nil
}

func initListener(port int, appLogger *slog.Logger) (net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		appLogger.Error("failed to start listener", "port", port, "error", err)
		return nil, fmt.Errorf("failed to listen port %d: %w", port, err)
	}
	appLogger.Info("net listener established", "addr", lis.Addr().String())
	return lis, nil
}

func initGameConn(ctx context.Context, cfg *config.Config, appLogger *slog.Logger) (*grpc.ClientConn, error) {
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
		appLogger.Error("failed to create gRPC client", "target", target, "error", err)
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	appLogger.Info("game client initialized successfully", "target", target)

	return conn, nil
}

func initGRPCServer(appLogger *slog.Logger) *grpc.Server {
	validator, _ := protovalidate.New()
	return grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			sharedinterceptor.LoggingInterceptor(appLogger),
			grpcadapter.ErrorInterceptor(),
			sharedinterceptor.ValidationUnaryInterceptor(validator),
		),
	)
}

func (a *app) Run(ctx context.Context) error {
	a.Logger.Info("starting gRPC service and workers...", "addr", a.listner.Addr().String())
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := a.server.Serve(a.listner); err != nil {
			return fmt.Errorf("gRPC server failed: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		a.Logger.Info("matchmaking worker started")
		a.matchmakingWorker.Start(ctx)
		return nil
	})
	return g.Wait()
}

func (a *app) Stop(ctx context.Context) {
	a.Logger.Info("graceful shutdown initiated for all components")

	stopped := make(chan struct{})
	go func() {
		a.Logger.Info("stopping public gRPC server...")
		a.server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		a.Logger.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		a.Logger.Warn("wait timeout exceeded, forcing gRPC server stop")
		a.server.Stop()
	}

	a.Logger.Info("closing redis client...")
	if err := a.redisClient.Close(); err != nil {
		a.Logger.Error("failed to close redis client", "error", err)
	} else {
		a.Logger.Info("redis client closed")
	}

	if a.shutdownOtel != nil {
		a.Logger.Info("shutting down telemetry...")
		if err := a.shutdownOtel(ctx); err != nil {
			a.Logger.Error("failed to shutdown telemetry gracefully", "error", err)
		} else {
			a.Logger.Info("telemetry shutdown successfully")
		}
	}

	a.Logger.Info("closing game client...")
	if err := a.gameConn.Close(); err != nil {
		a.Logger.Error("failed to close game client", "error", err)
	} else {
		a.Logger.Info("game client closed")
	}

	a.Logger.Info("application stopped completely")
}
