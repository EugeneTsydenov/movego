package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"shared/auth"
	"shared/telemetry"
	"user/internal/adapters/jwt"
	"user/internal/adapters/postgres"
	"user/internal/adapters/postgres/sqlc"
	"user/internal/application"
	"user/internal/config"

	"buf.build/go/protovalidate"
	userv1 "gen/user/v1"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	sharedinterceptor "shared/interceptor"

	grpcadapter "user/internal/adapters/grpc"
)

type app struct {
	logger          *slog.Logger
	shutdownOtel    func(context.Context) error
	db              *pgxpool.Pool
	cfg             *config.Config
	publicListener  net.Listener
	privateListener net.Listener
	publicServer    *grpc.Server
	privateServer   *grpc.Server
}

func newApp(ctx context.Context, logger *slog.Logger, cfg *config.Config, env string) (*app, error) {
	shutdownOtel, err := telemetry.Init(ctx, cfg.App.Name, cfg.Otel.Endpoint, cfg.Otel.MetricsPort)
	if err != nil {
		return nil, err
	}

	db, err := initDB(ctx, cfg)
	if err != nil {
		return nil, err
	}

	publicLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	privateLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.PrivatePort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	validator, _ := protovalidate.New()

	publicServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			sharedinterceptor.RecoveryInterceptor(logger),
			sharedinterceptor.LoggingInterceptor(logger),
			grpcadapter.ErrorInterceptor(),
			sharedinterceptor.ValidationUnaryInterceptor(validator),
		),
	)

	privateServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			sharedinterceptor.RecoveryInterceptor(logger),
			sharedinterceptor.LoggingInterceptor(logger),
			grpcadapter.ErrorInterceptor(),
			sharedinterceptor.ValidationUnaryInterceptor(validator),
			auth.ContextInterceptor(),
		),
	)

	reflection.Register(publicServer)
	reflection.Register(privateServer)

	return &app{
		logger:          logger,
		shutdownOtel:    shutdownOtel,
		db:              db,
		cfg:             cfg,
		publicListener:  publicLis,
		privateListener: privateLis,
		publicServer:    publicServer,
		privateServer:   privateServer,
	}, nil
}

func initDB(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
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

	return db, nil
}

func (a *app) InitDeps() error {
	querier := sqlc.New(a.db)
	unitOfWork := postgres.NewUnitOfWork(a.db)
	credentialRepo := postgres.NewCredentialRepo(querier)
	sessionRepo := postgres.NewSessionRepo(querier)

	tokenIssuer := jwt.NewIssuer(
		[]byte(os.Getenv("USER_SERVICE_JWT_SECRET_KEY")),
		a.cfg.JWT.AccessTTL,
		a.cfg.JWT.Issuer,
	)

	authService := application.NewAuthService(
		unitOfWork,
		credentialRepo,
		sessionRepo,
		tokenIssuer,
		a.cfg.JWT.RefreshTTL,
		a.logger,
	)
	sessionService := application.NewSessionService(sessionRepo)

	authHandler := grpcadapter.NewAuthHandler(authService)
	sessionHandler := grpcadapter.NewSessionHandler(sessionService)

	userv1.RegisterAuthServiceServer(a.publicServer, authHandler)
	userv1.RegisterSessionServiceServer(a.privateServer, sessionHandler)

	return nil
}

func (a *app) Run() error {
	g := new(errgroup.Group)

	g.Go(func() error {
		return a.publicServer.Serve(a.publicListener)
	})

	g.Go(func() error {
		return a.privateServer.Serve(a.privateListener)
	})

	return g.Wait()
}

func (a *app) Shutdown(ctx context.Context) {
	stopped := make(chan struct{})
	go func() {
		if a.publicServer != nil {
			a.publicServer.GracefulStop()
		}
		if a.privateServer != nil {
			a.privateServer.GracefulStop()
		}
		close(stopped)
	}()

	select {
	case <-stopped:
		a.logger.Info("gRPC servers gracefully stopped")
	case <-ctx.Done():
		a.logger.Warn("shutdown timeout")
		if a.publicServer != nil {
			a.publicServer.Stop()
		}
		if a.privateServer != nil {
			a.privateServer.Stop()
		}
	}

	if a.shutdownOtel != nil {
		if err := a.shutdownOtel(ctx); err != nil {
			a.logger.Error("failed to shutdown otel", "error", err)
		}
	}

	if a.db != nil {
		a.db.Close()
	}

	if a.publicListener != nil {
		if err := a.publicListener.Close(); err != nil {
			a.logger.Error("failed to close public listener", "error", err)
		}
	}

	if a.privateListener != nil {
		if err := a.privateListener.Close(); err != nil {
			a.logger.Error("failed to close private listener", "error", err)
		}
	}
}
