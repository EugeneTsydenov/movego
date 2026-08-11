package main

import (
	"context"
	"fmt"
	"log"
	movegov1 "movego/gen/go/movego/v1"
	grpcadapter "movego/internal/adapters/grpc"
	"movego/internal/adapters/jwt"
	"movego/internal/adapters/postgres"
	"movego/internal/application"
	"movego/internal/config"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type app struct {
	grpcServer *grpc.Server
	listener   net.Listener
	db         *pgxpool.Pool
}

func newApp(ctx context.Context, cfg *config.Config) (*app, error) {
	log.Print(cfg.Database)
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password='%s' dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.Server.Port, err)
	}

	grpcServer := grpc.NewServer()

	unitOfWork := postgres.NewUnitOfWork(db)
	tokenIssuer := jwt.NewIssuer([]byte(os.Getenv("MOVEGO_JWT_SECRET_KEY")), cfg.JWT.AccessTTL, cfg.JWT.Issuer)
	authServer := application.NewAuthService(unitOfWork, tokenIssuer, cfg.JWT.RefreshTTL)
	authHandler := grpcadapter.NewAuthHandler(authServer)

	movegov1.RegisterAuthServiceServer(grpcServer, authHandler)
	reflection.Register(grpcServer)

	return &app{
		grpcServer: grpcServer,
		listener:   lis,
		db:         db,
	}, nil
}

func (a *app) Run() error {
	return a.grpcServer.Serve(a.listener)
}

func (a *app) Stop(ctx context.Context) {
	a.db.Close()
	a.grpcServer.GracefulStop()
}
