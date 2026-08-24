package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"notification/internal/adapters/ws"
	"notification/internal/config"

	"shared/logger"
	"shared/telemetry"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"golang.org/x/sync/errgroup"
)

type app struct {
	server       *http.Server
	listner      net.Listener
	shutdownOtel func(context.Context) error
	Logger       *slog.Logger
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

	wsManager := ws.NewManager()
	wsHandler := ws.NewHandler(wsManager, appLogger)

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
		a.server.Shutdown(ctx)
		close(stopped)
	}()

	select {
	case <-stopped:
		a.Logger.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		a.Logger.Warn("wait timeout exceeded, forcing gRPC server stop")
		a.server.Close()
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
