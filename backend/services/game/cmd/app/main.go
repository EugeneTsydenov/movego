package main

import (
	"context"
	"flag"
	"game/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"shared/logger"
	"syscall"
)

const envPrefix = "GAME_SERVICE"

func fetchAppEnv() string {
	return os.Getenv("GAME_SERVICE_ENV")
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
	env := fetchAppEnv()
	configDir := fetchConfigDir(envPrefix)
	cfg, err := config.Load(configDir, env, envPrefix)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	appLogger := logger.New(env, logger.FromStringLevel(cfg.App.LogLevel))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	appLogger.Info("initializing game service app...", "env", env)
	app, err := newApp(ctx, appLogger, cfg, env)
	if err != nil {
		appLogger.Error("failed to init app", "error", err)
		os.Exit(1)
	}
	appLogger.Info("app initialized")

	appLogger.Info("initializing deps...")
	app.InitDeps()
	appLogger.Info("deps initialized")

	go func() {
		appLogger.Info("running app...")
		if err := app.Run(); err != nil {
			appLogger.Error("application runtime error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	appLogger.Info("game service shutting down...")
	app.Shutdown(shutdownCtx)
	appLogger.Info("game service stopped")
}
