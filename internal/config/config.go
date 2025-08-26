package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

type Config struct {
	Logging   LoggingConfig   `mapstructure:"logging"`
	Telemetry TelemetryConfig `mapstructure:"telemetry"`
	Server    ServerConfig    `mapstructure:"server"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type TelemetryConfig struct {
	ServiceName    string `mapstructure:"service_name"`
	ServiceVersion string `mapstructure:"service_version"`
	Enabled        bool   `mapstructure:"enabled"`
}

var configFileNames = []string{"default", "local", "private"}

func Load() (*Config, error) {
	k := koanf.New(".")
	// Load base files in desired sequence
	if err := k.Load(file.Provider("./configs/default.yaml"), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}
	if err := k.Load(file.Provider("./configs/local.yaml"), yaml.Parser()); err != nil {
		// ignore “missing local” if you want
	}
	if err := k.Load(file.Provider("./configs/private.yaml"), yaml.Parser()); err != nil {
		// ignore “missing private” if you want
	}
	// Load environment variables
	err := k.Load(env.Provider("APP_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), "_", ".")
	}), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to bind environment variables: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &cfg, nil
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
