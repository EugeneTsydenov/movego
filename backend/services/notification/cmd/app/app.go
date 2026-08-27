package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	natsadapter "notification/internal/adapters/nats"
	"notification/internal/adapters/ws"
	"notification/internal/config"

	"shared/logger"
	"shared/otelnats"
	"shared/telemetry"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/sync/errgroup"
)

type app struct {
	Logger       *slog.Logger
	shutdownOtel func(context.Context) error
	listner      net.Listener
	server       *http.Server
	natsConn     *nats.Conn
	natsSub      *natsadapter.GameJetStreamSub
}

func newApp(ctx context.Context, cfg *config.Config, env string) (*app, error) {
	appLogger := logger.New(env, logger.FromStringLevel(cfg.App.LogLevel))

	shutdownOtel, err := telemetry.Init(ctx, cfg.App.Name, cfg.Otel.Endpoint, cfg.Otel.MetricsPort)
	if err != nil {
		return nil, err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
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

	wsManager := ws.NewManager()
	wsHandler := ws.NewHandler(wsManager, appLogger)

	gameHandler := natsadapter.NewGameHandler(wsManager, appLogger)

	natsSub := natsadapter.NewGameJetStreamSub(
		js,
		otelnats.TraceMiddleware(cfg.App.Name, appLogger, gameHandler.Route),
		appLogger,
	)

	mux := http.NewServeMux()
	wrappedHandler := otelhttp.NewHandler(mux, "ws-server")
	mux.Handle("/ws", wsHandler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	httpServer := &http.Server{
		Handler: wrappedHandler,
	}

	return &app{
		server:       httpServer,
		listner:      lis,
		shutdownOtel: shutdownOtel,
		Logger:       appLogger,
		natsConn:     nc,
		natsSub:      natsSub,
	}, nil
}

func (a *app) Run() error {
	g := new(errgroup.Group)

	g.Go(func() error {
		return a.server.Serve(a.listner)
	})

	g.Go(func() error {
		return a.natsSub.Start(context.Background())
	})

	return g.Wait()
}

func (a *app) Shutdown(ctx context.Context) {
	stopped := make(chan struct{})
	go func() {
		a.server.Shutdown(ctx)
		a.natsSub.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
		a.Logger.Info("http server gracefully stopped")
	case <-ctx.Done():
		a.Logger.Warn("shutdown timeout")
		a.server.Close()
	}

	if a.natsConn != nil {
		a.natsConn.Close()
	}

	if a.listner != nil {
		if err := a.listner.Close(); err != nil {
			a.Logger.Error("failed to close listener", "error", err)
		}
	}

	if a.shutdownOtel != nil {
		if err := a.shutdownOtel(context.Background()); err != nil {
			a.Logger.Error("failed to shutdown otel", "error", err)
		}
	}
}
