package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	gamev1 "protogen/game/v1"

	grpcadapter "game/internal/adapters/grpc"
	redisadapter "game/internal/adapters/redis"
	"game/internal/application"
	"game/internal/config"

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
	"google.golang.org/grpc/reflection"
)

type app struct {
	server       *grpc.Server
	listner      net.Listener
	redisClient  *redis.Client
	shutdownOtel func(context.Context) error
	Logger       *slog.Logger
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

	initGameModule(cfg, server, redisClient, appLogger)

	reflection.Register(server)

	appLogger.Info("gRPC services registered successfully")

	return &app{
		server:       server,
		listner:      lis,
		redisClient:  redisClient,
		shutdownOtel: shutdownOtel,
		Logger:       appLogger,
	}, nil
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

func initGameModule(cfg *config.Config, server *grpc.Server, redisClient *redis.Client, appLogger *slog.Logger) {
	gameRepository := redisadapter.NewGameRepository(redisClient)
	gameService := application.NewGameService(gameRepository)
	gamev1.RegisterGameServiceServer(server, grpcadapter.NewGameHandler(gameService))
}

func (a *app) Run() error {
	a.Logger.Info("starting gRPC service...", "addr", a.listner.Addr().String())
	g := new(errgroup.Group)
	g.Go(func() error {
		if err := a.server.Serve(a.listner); err != nil {
			return fmt.Errorf("gRPC server failed: %w", err)
		}
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

	a.Logger.Info("application stopped completely")
}
