package main

import (
	"context"
	"flag"
	"log/slog"
	"notification/internal/config"
	"os"
	"os/signal"
	"shared/logger"
	"syscall"
)

const envPrefix = "NOTIFICATION_SERVICE"

func fetchAppEnv() string {
	return os.Getenv("NOTIFICATION_SERVICE_ENV")
}

func fetchConfigDir(prefix string) string {
	var res string

	flag.StringVar(&res, "config", "", "path to config dir")
	flag.StringVar(&res, "c", "", "path to config dir(shorter)")
	flag.Parse()

	if res != "" {
		return res
	}

	if res = os.Getenv(prefix + "_" + "CONFIG_PATH"); res != "" {
		return res
	}

	return "configs"
}

func main() {
	configDir := fetchConfigDir(envPrefix)
	env := fetchAppEnv()
	cfg, err := config.Load(configDir, env, envPrefix)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	appLogger := logger.New(env, logger.FromStringLevel(cfg.App.LogLevel))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	appLogger.Info("initializing notification service app", "env", env)
	app, err := newApp(ctx, appLogger, cfg, env)
	if err != nil {
		appLogger.Error("failed to init app", "error", err)
		os.Exit(1)
	}
	appLogger.Info("app initialized")

	go func() {
		appLogger.Info("app running")
		if err := app.Run(); err != nil {
			appLogger.Error("application runtime error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	appLogger.Info("notification service shutting down")
	app.Shutdown(shutdownCtx)
	appLogger.Info("notification service stopped")
}
