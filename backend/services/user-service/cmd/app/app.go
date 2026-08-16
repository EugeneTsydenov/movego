package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"

	userv1 "protogen/user/v1"
	grpcadapter "user/internal/adapters/grpc"
	"user/internal/adapters/jwt"
	"user/internal/adapters/postgres"
	"user/internal/adapters/postgres/sqlc"
	"user/internal/application"
	"user/internal/config"
	"user/internal/pkg/logger"
	"user/internal/pkg/telemetry"

	"buf.build/go/protovalidate"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type app struct {
	publicServer    *grpc.Server
	publicListener  net.Listener
	privateServer   *grpc.Server
	privateListener net.Listener
	db              *pgxpool.Pool
	shutdownOtel    func(context.Context) error
	Logger          *slog.Logger
}

func newApp(ctx context.Context, cfg *config.Config, env string) (*app, error) {
	appLogger := initLogger(cfg, env)

	shutdownOtel, err := initTelemetry(ctx, cfg, appLogger)
	if err != nil {
		return nil, err
	}

	db, err := initDatabase(ctx, cfg, appLogger)
	if err != nil {
		return nil, err
	}

	publicLis, err := initListener(cfg.Server.Port, "public", appLogger)
	if err != nil {
		return nil, err
	}
	publicServer := initPublicGRPCServer(appLogger)

	privateLis, err := initListener(cfg.Server.PrivatePort, "private", appLogger)
	if err != nil {
		return nil, err
	}
	privateServer := initPrivateGRPCServer(appLogger)

	initAuthModule(publicServer, privateServer, db, cfg, appLogger)

	reflection.Register(publicServer)
	reflection.Register(privateServer)
	appLogger.Info("gRPC services registered successfully on both servers")

	return &app{
		publicServer:    publicServer,
		publicListener:  publicLis,
		privateServer:   privateServer,
		privateListener: privateLis,
		db:              db,
		shutdownOtel:    shutdownOtel,
		Logger:          appLogger,
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

func initDatabase(ctx context.Context, cfg *config.Config, appLogger *slog.Logger) (*pgxpool.Pool, error) {
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
	return db, nil
}

func initListener(port int, serverType string, appLogger *slog.Logger) (net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		appLogger.Error("failed to start listener", "type", serverType, "port", port, "error", err)
		return nil, fmt.Errorf("failed to listen on %s port %d: %w", serverType, port, err)
	}
	appLogger.Info("net listener established", "type", serverType, "addr", lis.Addr().String())
	return lis, nil
}

func initPublicGRPCServer(appLogger *slog.Logger) *grpc.Server {
	validator, _ := protovalidate.New()

	return grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			grpcadapter.LoggingInterceptor(appLogger),
			grpcadapter.ErrorInterceptor(),
			grpcadapter.ValidationUnaryInterceptor(validator),
		),
	)
}

func initPrivateGRPCServer(appLogger *slog.Logger) *grpc.Server {
	validator, _ := protovalidate.New()

	return grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			grpcadapter.LoggingInterceptor(appLogger),
			grpcadapter.ErrorInterceptor(),
			grpcadapter.ValidationUnaryInterceptor(validator),
			grpcadapter.AuthContextInterceptor(),
		),
	)
}

func initAuthModule(grpcServer, privateGRPCServer *grpc.Server, db *pgxpool.Pool, cfg *config.Config, appLogger *slog.Logger) {
	unitOfWork := postgres.NewUnitOfWork(db)
	querier := sqlc.New(db)
	credentialRepo := postgres.NewCredentialRepo(querier)
	sessionRepo := postgres.NewSessionRepo(querier)

	tokenIssuer := jwt.NewIssuer([]byte(os.Getenv("user_JWT_SECRET_KEY")), cfg.JWT.AccessTTL, cfg.JWT.Issuer)

	authServer := application.NewAuthService(
		unitOfWork,
		credentialRepo,
		sessionRepo,
		tokenIssuer,
		cfg.JWT.RefreshTTL,
		appLogger,
	)
	sessionService := application.NewSessionService(sessionRepo)

	authHandler := grpcadapter.NewAuthHandler(authServer)
	sessionHandler := grpcadapter.NewSessionHandler(sessionService)
	userv1.RegisterAuthServiceServer(grpcServer, authHandler)
	userv1.RegisterSessionServiceServer(privateGRPCServer, sessionHandler)
}

func (a *app) Run() error {
	a.Logger.Info("starting gRPC servers...", "public_addr", a.publicListener.Addr().String(), "private_addr", a.privateListener.Addr().String())

	g := new(errgroup.Group)

	g.Go(func() error {
		if err := a.publicServer.Serve(a.publicListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("public gRPC server failed: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := a.privateServer.Serve(a.privateListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("private gRPC server failed: %w", err)
		}
		return nil
	})

	return g.Wait()
}

func (a *app) Stop(ctx context.Context) {
	a.Logger.Info("graceful shutdown initiated for all components")

	stoppedPublic := make(chan struct{})
	go func() {
		a.Logger.Info("stopping public gRPC server...")
		a.publicServer.GracefulStop()
		close(stoppedPublic)
	}()

	stoppedPrivate := make(chan struct{})
	go func() {
		a.Logger.Info("stopping private gRPC server...")
		a.privateServer.GracefulStop()
		close(stoppedPrivate)
	}()

	select {
	case <-stoppedPublic:
		<-stoppedPrivate
		a.Logger.Info("both gRPC servers stopped gracefully")
	case <-ctx.Done():
		a.Logger.Warn("wait timeout exceeded, forcing gRPC servers stop")
		a.publicServer.Stop()
		a.privateServer.Stop()
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
