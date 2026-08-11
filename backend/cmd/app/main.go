package main

import (
	"context"
	"flag"
	"log"
	"movego/internal/config"
	"os"
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
	cfg, err := config.Load(configDir, os.Getenv("APP_ENV"), prefix)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	app, err := newApp(context.TODO(), cfg)
	if err := app.Run(); err != nil {
		log.Fatalf("failed to run grpc server: %v", err)
	}
}
