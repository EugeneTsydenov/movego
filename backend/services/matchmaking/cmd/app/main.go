package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"matchmaking/internal/config"
)

const envPrefix = "MATCHMAKING_SERVICE"

func fetchAppEnv() string {
	return os.Getenv("MATCHMAKING_SERVICE_ENV")
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
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("initializing matchmaking service app in %s mode...", env)
	app, err := newApp(ctx, cfg, env)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}
	log.Print("app initialized")

	log.Print("initializing deps...")
	app.InitDeps()
	log.Print("deps initialized")

	go func() {
		log.Print("running app..")
		if err := app.Run(ctx); err != nil {
			log.Fatalf("application runtime error: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	log.Print("matchmaking service shutting down...")
	app.Shutdown(shutdownCtx)
	log.Print("matchmaking service stopped")
}
