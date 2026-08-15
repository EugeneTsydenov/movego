package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	movegov1 "movego/gen/go/movego/v1"
	grpcadapter "movego/internal/adapters/grpc"
	"movego/internal/adapters/jwt"
	"movego/internal/adapters/postgres"
	"movego/internal/adapters/postgres/sqlc"
	"movego/internal/application"
	"movego/internal/config"
	"movego/internal/pkg/logger"
	"movego/internal/pkg/telemetry"

	"buf.build/go/protovalidate"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type app struct {
	grpcServer   *grpc.Server
	listener     net.Listener
	db           *pgxpool.Pool
	shutdownOtel func(context.Context) error
	Logger       *slog.Logger
}

func newApp(ctx context.Context, cfg *config.Config, env string) (*app, error) {
	appLogger := logger.New(env, logger.FromStringLevel(cfg.App.LogLevel))
	appLogger.Info("initializing application", "app_name", cfg.App.Name, "env", env)

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		appLogger.Error("opentelemetry internal error", "error", err)
	}))

	shutdownOtel, err := telemetry.InitTelemetry(ctx, cfg.App.Name, cfg.Otel.Endpoint, cfg.Otel.MetricsPort)
	if err != nil {
		appLogger.Error("failed to init telemetry", "error", err)
		return nil, fmt.Errorf("failed to init telemetry: %w", err)
	}
	appLogger.Info("telemetry initialized", "endpoint", cfg.Otel.Endpoint, "metrics_port", cfg.Otel.MetricsPort)

	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password='%s' dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	pgCfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		appLogger.Error("failed to parse postgres config", "error", err)
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}
	pgCfg.ConnConfig.Tracer = otelpgx.NewTracer()

	db, err := pgxpool.NewWithConfig(ctx, pgCfg)
	if err != nil {
		appLogger.Error("failed to create pgx pool", "error", err)
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	if err = db.Ping(ctx); err != nil {
		appLogger.Error("failed to ping postgres", "host", cfg.Database.Host, "error", err)
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}
	appLogger.Info("connected to postgres successfully", "host", cfg.Database.Host, "database", cfg.Database.Name)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		appLogger.Error("failed to start listener", "port", cfg.Server.Port, "error", err)
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.Server.Port, err)
	}
	appLogger.Info("net listener established", "addr", lis.Addr().String())

	validator, _ := protovalidate.New()

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			grpcadapter.LoggingInterceptor(appLogger),
			grpcadapter.ErrorInterceptor(),
			grpcadapter.ValidationUnaryInterceptor(validator),
		),
	)

	unitOfWork := postgres.NewUnitOfWork(db)
	querier := sqlc.New(db)
	credentialRepo := postgres.NewCredentialRepo(querier)
	sessionRepo := postgres.NewSessionRepo(querier)
	tokenIssuer := jwt.NewIssuer([]byte(os.Getenv("MOVEGO_JWT_SECRET_KEY")), cfg.JWT.AccessTTL, cfg.JWT.Issuer)
	authServer := application.NewAuthService(unitOfWork, credentialRepo, sessionRepo, tokenIssuer, cfg.JWT.RefreshTTL)
	authHandler := grpcadapter.NewAuthHandler(authServer)

	movegov1.RegisterAuthServiceServer(grpcServer, authHandler)
	reflection.Register(grpcServer)
	appLogger.Info("gRPC services registered successfully")

	return &app{
		grpcServer:   grpcServer,
		listener:     lis,
		db:           db,
		shutdownOtel: shutdownOtel,
		Logger:       appLogger,
	}, nil
}

func (a *app) Run() error {
	a.Logger.Info("starting gRPC server...", "addr", a.listener.Addr().String())
	return a.grpcServer.Serve(a.listener)
}

func (a *app) Stop(ctx context.Context) {
	a.Logger.Info("graceful shutdown initiated")

	stopped := make(chan struct{})
	go func() {
		a.Logger.Info("stopping gRPC server (waiting for active RPCs to finish)...")
		a.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		a.Logger.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		a.Logger.Warn("wait timeout exceeded, forcing gRPC server stop")
		a.grpcServer.Stop()
	}

	a.Logger.Info("closing database connection pool...")
	a.db.Close()
	a.Logger.Info("database connection pool closed")

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
