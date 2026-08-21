package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func Load(configDir, appEnv, prefix string, cfg any) error {
	v := viper.New()

	v.SetEnvPrefix(prefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("database.password", "")
	v.AddConfigPath(configDir)
	v.SetConfigType("yaml")

	v.SetConfigName("config")
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read base config.yaml: %w", err)
	}

	if appEnv == "dev" || appEnv == "prod" {
		v.SetConfigName("config." + appEnv)
		if err := v.MergeInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				return fmt.Errorf("failed to merge %s config: %w", appEnv, err)
			}
		}
	}

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
