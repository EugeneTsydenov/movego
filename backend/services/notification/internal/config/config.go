package config

import (
	sharedconfig "shared/config"
	"time"
)

type Config struct {
	App struct {
		Name            string        `mapstructure:"name"`
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
		LogLevel        string        `mapstructure:"log_level"`
	} `mapstructure:"app"`

	Server struct {
		Host string `mapstructure:"host" json:"host"`
		Port int    `mapstructure:"port" json:"port"`
	} `mapstructure:"server"`

	Nats struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"nats"`

	Otel struct {
		Endpoint    string `mapstructure:"endpoint"`
		MetricsPort int    `mapstructure:"metrics_port"`
	} `mapstructure:"otel"`
}

func Load(configDir, appEnv, prefix string) (*Config, error) {
	var cfg Config
	if err := sharedconfig.Load(configDir, appEnv, prefix, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
