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

	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Name     string `mapstructure:"name"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Driver   string `mapstructure:"driver"`
		SSLMode  string `mapstructure:"ssl_mode"`
		MaxConn  int    `mapstructure:"max_conn"`
	} `mapstructure:"database"`

	Server struct {
		Host            string        `mapstructure:"host" json:"host"`
		Port            int           `mapstructure:"port" json:"port"`
		ReadTimeout     time.Duration `mapstructure:"read_timeout" json:"read_timeout"`
		WriteTimeout    time.Duration `mapstructure:"write_timeout" json:"write_timeout"`
		IdleTimeout     time.Duration `mapstructure:"idle_timeout" json:"idle_timeout"`
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout"`
	} `mapstructure:"server"`

	Otel struct {
		Endpoint    string `mapstructure:"endpoint"`
		MetricsPort int    `mapstructure:"metrics_port"`
	} `mapstructure:"otel"`
}

func Load(configDir, appEnv, prefix string) (*Config, error) {
	var cfg Config
	cfg.Database.Password = ""
	if err := sharedconfig.Load(configDir, appEnv, prefix, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
