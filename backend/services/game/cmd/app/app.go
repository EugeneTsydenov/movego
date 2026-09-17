package main

import (
	"context"
	"fmt"
	"game/internal/adapters/ws"
	"game/internal/application"
	"game/internal/config"
	"log/slog"
	"net"
	"net/http"
	"shared/otelnats"
	"shared/telemetry"

	grpcadapter "game/internal/adapters/grpc"
	natsadapter "game/internal/adapters/nats"
	redisadapter "game/internal/adapters/redis"

	"buf.build/go/protovalidate"

	gamev1 "gen/game/v1"
	sharedinterceptor "shared/interceptor"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	sharedredis "shared/redis"
)

type app struct {
	cfg    *config.Config
	logger *slog.Logger

	shutdownOtel func(context.Context) error
	redisClient  *redis.Client
	natsConn     *nats.Conn

	js jetstream.JetStream

	listener   net.Listener
	grpcServer *grpc.Server

	wsListener net.Listener
	wsServer   *http.Server
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

	nc, err := nats.Connect(cfg.Nats.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to init jetstream: %w", err)
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
		),
	)
	reflection.Register(grpcServer)

	wsListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.WebSocketServer.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen ws: %w", err)
	}
	wsServer := &http.Server{}

	return &app{
		cfg:          cfg,
		logger:       logger,
		shutdownOtel: shutdownOtel,
		redisClient:  redisClient,
		natsConn:     nc,
		js:           js,
		listener:     lis,
		grpcServer:   grpcServer,
		wsListener:   wsListener,
		wsServer:     wsServer,
	}, nil
}

func (a *app) InitDeps() {
	gameRepo := redisadapter.NewGameRepo(a.redisClient)

	basePublish := func(ctx context.Context, msg *nats.Msg, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error) {
		return a.js.PublishMsg(ctx, msg, opts...)
	}
	tracedPublish := otelnats.TraceInjectMiddleware(a.cfg.App.Name, basePublish)
	gamePublisher := natsadapter.NewGamePublisher(tracedPublish, a.logger)

	gameService := application.NewGameService(gameRepo, gamePublisher)

	gameHandler := grpcadapter.NewGameHandler(gameService)
	gamev1.RegisterGameServiceServer(a.grpcServer, gameHandler)

	wsManager := ws.NewGameManager()
	wsHandler := ws.NewGameHandler(wsManager, gameService, a.logger)

	mux := http.NewServeMux()
	wsHandler.Handle(mux)
	a.wsServer.Handler = otelhttp.NewHandler(mux, "ws-game-server")
}

func (a *app) Run() error {
	g := new(errgroup.Group)
	g.Go(func() error {
		return a.grpcServer.Serve(a.listener)
	})
	g.Go(func() error {
		return a.wsServer.Serve(a.wsListener)
	})
	return g.Wait()
}

func (a *app) Shutdown(ctx context.Context) {
	stopped := make(chan struct{})
	go func() {
		if a.grpcServer != nil {
			a.grpcServer.GracefulStop()
		}
		if a.wsServer != nil {
			a.wsServer.Shutdown(ctx)
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
		if a.wsServer != nil {
			a.wsServer.Close()
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
