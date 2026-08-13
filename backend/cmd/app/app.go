package main

import (
	"context"
	"fmt"
	movegov1 "movego/gen/go/movego/v1"
	grpcadapter "movego/internal/adapters/grpc"
	"movego/internal/adapters/jwt"
	"movego/internal/adapters/postgres"
	"movego/internal/application"
	"movego/internal/config"
	"movego/internal/pkg/logger"
	"movego/internal/pkg/telemetry"
	"net"
	"os"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type app struct {
	grpcServer   *grpc.Server
	listener     net.Listener
	db           *pgxpool.Pool
	shutdownOtel func(context.Context) error
}

func newApp(ctx context.Context, cfg *config.Config, env string) (*app, error) {
	shutdownOtel, err := telemetry.InitTelemetry(ctx, cfg.App.Name, cfg.Otel.Endpoint, cfg.Otel.MetricsPort)
	if err != nil {
		return nil, fmt.Errorf("failed to init telemetry: %w", err)
	}

	appLogger := logger.New(env, logger.FromStringLevel(cfg.App.LogLevel))

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
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}
	pgCfg.ConnConfig.Tracer = otelpgx.NewTracer()

	db, err := pgxpool.NewWithConfig(ctx, pgCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	if err = db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.Server.Port, err)
	}

	//interceptors
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			grpcadapter.ErrorHInterceptor(),
			grpcadapter.LoggingInterceptor(appLogger),
		),
	)

	unitOfWork := postgres.NewUnitOfWork(db)
	tokenIssuer := jwt.NewIssuer([]byte(os.Getenv("MOVEGO_JWT_SECRET_KEY")), cfg.JWT.AccessTTL, cfg.JWT.Issuer)
	authServer := application.NewAuthService(unitOfWork, tokenIssuer, cfg.JWT.RefreshTTL)
	authHandler := grpcadapter.NewAuthHandler(authServer)

	movegov1.RegisterAuthServiceServer(grpcServer, authHandler)
	reflection.Register(grpcServer)

	return &app{
		grpcServer:   grpcServer,
		listener:     lis,
		db:           db,
		shutdownOtel: shutdownOtel,
	}, nil
}

func (a *app) Run() error {
	return a.grpcServer.Serve(a.listener)
}

func (a *app) Stop(ctx context.Context) {
	a.db.Close()
	a.grpcServer.GracefulStop()
	if a.shutdownOtel != nil {
		a.shutdownOtel(ctx)
	}
}
