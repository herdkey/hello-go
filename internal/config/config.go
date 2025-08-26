package config

import (
	"errors"
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

var configFileNames = []string{"default", "local", "private"}

func Load() (*Config, error) {
	v := viper.New()

	config, err := mergeConfigFiles(v)
	if err != nil {
		return config, err
	}

	if err := bindEnv(v); err != nil {
		return nil, fmt.Errorf("failed to bind environment variables: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func mergeConfigFiles(v *viper.Viper) (*Config, error) {
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")
	v.SetConfigType("yaml")

	// Merge in config files
	for _, configName := range configFileNames {
		v.SetConfigName(configName)
		if err := v.MergeInConfig(); err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return nil, fmt.Errorf("failed to merge %s config: %w", configName, err)
		}
	}
	return nil, nil
}

func bindEnv(v *viper.Viper) error {
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	bindings := map[string]string{
		"server.host":       "APP_SERVER_HOST",
		"server.port":       "APP_SERVER_PORT",
		"logging.level":     "APP_LOGGING_LEVEL",
		"logging.format":    "APP_LOGGING_FORMAT",
		"telemetry.enabled": "APP_TELEMETRY_ENABLED",
	}

	// Bind environment variables explicitly for nested keys
	for key, env := range bindings {
		if err := v.BindEnv(key, env); err != nil {
			return err
		}
	}
	return nil
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
