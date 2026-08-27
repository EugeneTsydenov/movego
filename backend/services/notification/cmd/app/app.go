package main

import (
	"context"
	"errors"
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
	"go.opentelemetry.io/otel"
	"golang.org/x/sync/errgroup"
)

type app struct {
	server       *http.Server
	listner      net.Listener
	shutdownOtel func(context.Context) error
	Logger       *slog.Logger
	natsConn     *nats.Conn
	natsSub      *natsadapter.GameJetStreamSub
}

func newApp(ctx context.Context, cfg *config.Config, env string) (*app, error) {
	appLogger := initLogger(cfg, env)

	shutdownOtel, err := initTelemetry(ctx, cfg, appLogger)
	if err != nil {
		return nil, err
	}

	lis, err := initListener(cfg.Server.Port, appLogger)
	if err != nil {
		return nil, err
	}

	nc, err := nats.Connect(cfg.Nats.URL)
	if err != nil {
		appLogger.Error("failed to connect to nats", "error", err)
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		appLogger.Error("failed to init jetstream", "error", err)
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

	appLogger.Info("services registered successfully")

	return &app{
		server:       httpServer,
		listner:      lis,
		shutdownOtel: shutdownOtel,
		Logger:       appLogger,
		natsConn:     nc,
		natsSub:      natsSub,
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

func initListener(port int, appLogger *slog.Logger) (net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		appLogger.Error("failed to start listener", "port", port, "error", err)
		return nil, fmt.Errorf("failed to listen port %d: %w", port, err)
	}
	appLogger.Info("net listener established", "addr", lis.Addr().String())
	return lis, nil
}

func (a *app) Run() error {
	a.Logger.Info("starting application services...", "addr", a.listner.Addr().String())
	g := new(errgroup.Group)

	g.Go(func() error {
		if err := a.server.Serve(a.listner); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server failed: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := a.natsSub.Start(context.Background()); err != nil {
			return fmt.Errorf("nats consumer failed: %w", err)
		}
		return nil
	})

	return g.Wait()
}

func (a *app) Stop(ctx context.Context) {
	a.Logger.Info("graceful shutdown initiated for all components")

	a.Logger.Info("stopping nats consumer...")
	a.natsSub.Stop()

	stopped := make(chan struct{})
	go func() {
		a.Logger.Info("stopping http server...")
		a.server.Shutdown(ctx)
		close(stopped)
	}()

	select {
	case <-stopped:
		a.Logger.Info("http server stopped gracefully")
	case <-ctx.Done():
		a.Logger.Warn("wait timeout exceeded, forcing http server stop")
		a.server.Close()
	}

	if a.natsConn != nil {
		a.Logger.Info("closing nats connection...")
		a.natsConn.Close()
	}

	if a.shutdownOtel != nil {
		a.Logger.Info("shutting down telemetry...")
		if err := a.shutdownOtel(context.Background()); err != nil {
			a.Logger.Error("failed to shutdown telemetry gracefully", "error", err)
		} else {
			a.Logger.Info("telemetry shutdown successfully")
		}
	}

	a.Logger.Info("application stopped completely")
}
