package main

import (
	"context"
	"flag"
	"log"
	"movego/internal/config"
	"os"
	"os/signal"
	"syscall"
)

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
	prefix := "MOVEGO"
	configDir := fetchConfigDir(prefix)
	env := os.Getenv("APP_ENV")
	cfg, err := config.Load(configDir, env, prefix)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := newApp(ctx, cfg, env)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}

	go func() {
		if err := app.Run(); err != nil {
			app.Logger.Error("grpc server runtime error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	app.Stop(shutdownCtx)
}
