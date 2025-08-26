package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
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

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config.default")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	// Bind environment variables explicitly for nested keys
	_ = v.BindEnv("server.host", "APP_SERVER_HOST")
	_ = v.BindEnv("server.port", "APP_SERVER_PORT")
	_ = v.BindEnv("logging.level", "APP_LOGGING_LEVEL")
	_ = v.BindEnv("logging.format", "APP_LOGGING_FORMAT")
	_ = v.BindEnv("telemetry.enabled", "APP_TELEMETRY_ENABLED")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
