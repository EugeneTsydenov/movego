package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name     string `mapstructure:"name"`
		LogLevel string `mapstructure:"log_level"`
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

	JWT struct {
		RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
		AccessTTL  time.Duration `mapstructure:"access_ttl"`
		Issuer     string        `mapstructure:"issuer"`
	} `mapstructure:"jwt"`

	Otel struct {
		Endpoint    string `mapstructure:"endpoint"`
		MetricsPort int    `mapstructure:"metrics_port"`
	} `mapstructure:"otel"`
}

func Load(configDir, appEnv, prefix string) (*Config, error) {
	v := viper.New()

	v.SetEnvPrefix(prefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("database.password", "")
	v.AddConfigPath(configDir)
	v.SetConfigType("yaml")

	v.SetConfigName("config")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read base config.yaml: %w", err)
	}

	if appEnv == "dev" || appEnv == "prod" {
		v.SetConfigName("config." + appEnv)
		if err := v.MergeInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				return nil, fmt.Errorf("failed to merge %s config: %w", appEnv, err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
