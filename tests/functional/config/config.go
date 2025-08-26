package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()

	// Set default config
	v.SetConfigName("default")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading default config: %w", err)
	}

	// Merge local config
	v.SetConfigName("local")
	if err := v.MergeInConfig(); err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
		return nil, fmt.Errorf("failed to merge local config: %w", err)
	}

	// Merge private config
	v.SetConfigName("private")
	if err := v.MergeInConfig(); err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
		return nil, fmt.Errorf("failed to merge private config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &cfg, nil
}
