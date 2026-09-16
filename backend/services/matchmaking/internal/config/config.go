package config

import (
	"time"

	sharedconfig "shared/config"
)

type Config struct {
	App struct {
		Name            string        `mapstructure:"name"`
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
		LogLevel        string        `mapstructure:"log_level"`
	} `mapstructure:"app"`

	Redis struct {
		Addr            string        `mapstructure:"addr"`
		Username        string        `mapstructure:"username"`
		Password        string        `mapstructure:"password"`
		DB              int           `mapstructure:"db"`
		PoolSize        int           `mapstructure:"pool_size"`
		MinIdleConns    int           `mapstructure:"min_idle_conns"`
		ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
		DialTimeout     time.Duration `mapstructure:"dial_timeout"`
		ReadTimeout     time.Duration `mapstructure:"read_timeout"`
		WriteTimeout    time.Duration `mapstructure:"write_timeout"`
		MaxRetries      int           `mapstructure:"max_retries"`
	} `mapstructure:"redis"`

	Server struct {
		Host            string        `mapstructure:"host" json:"host"`
		Port            int           `mapstructure:"port" json:"port"`
		ReadTimeout     time.Duration `mapstructure:"read_timeout" json:"read_timeout"`
		WriteTimeout    time.Duration `mapstructure:"write_timeout" json:"write_timeout"`
		IdleTimeout     time.Duration `mapstructure:"idle_timeout" json:"idle_timeout"`
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout"`
	} `mapstructure:"server"`

	GameClient struct {
		Host        string        `mapstructure:"host" json:"host"`
		Port        int           `mapstructure:"port" json:"port"`
		DialTimeout time.Duration `mapstructure:"dial_timeout" json:"dial_timeout"`
		ReadTimeout time.Duration `mapstructure:"read_timeout" json:"read_timeout"`
		MaxRetries  int           `mapstructure:"max_retries" json:"max_retries"`
		TlsEnabled  bool          `mapstructure:"tls_enabled" json:"tls_enabled"`
	} `mapstructure:"game_client"`

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
	cfg.Redis.Password = ""
	if err := sharedconfig.Load(configDir, appEnv, prefix, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
