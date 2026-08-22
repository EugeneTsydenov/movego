package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"{{cookiecutter.service_name}}/internal/config"
)

const envPrefix = "{{cookiecutter.service_name | upper}}_SERVICE"

func fetchAppEnv() string {
	return os.Getenv("{{cookiecutter.service_name | upper}}_SERVICE_ENV")
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

	app, err := newApp(ctx, cfg, env)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}

	go func() {
		if err := app.Run(); err != nil {
			app.Logger.Error("application runtime error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	app.Stop(shutdownCtx)
}
