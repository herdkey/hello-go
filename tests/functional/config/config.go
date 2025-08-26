package config

import (
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
	v.AddConfigPath("./tests/functional/config")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading default config: %w", err)
	}

	// Merge local config
	v.SetConfigName("local")
	if err := v.MergeInConfig(); err != nil {
		// Local config might not exist; proceed if it doesn't
	}

	// Merge private config
	v.SetConfigName("private")
	if err := v.MergeInConfig(); err != nil {
		// Private config might not exist; proceed if it doesn't
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &cfg, nil
}
