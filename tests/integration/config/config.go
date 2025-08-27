package config

import (
	"fmt"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func (s *ServerConfig) URL() string {
    return fmt.Sprintf("http://%s:%d", s.Host, s.Port)
}

func LoadConfig() (*Config, error) {
	k := koanf.New(".")
	if err := k.Load(file.Provider("./config/default.yaml"), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("error reading default config: %w", err)
	}
	// Attempt local merge
	_ = k.Load(file.Provider("./config/local.yaml"), yaml.Parser())
	// Attempt private merge
	_ = k.Load(file.Provider("./config/private.yaml"), yaml.Parser())

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}
	return &cfg, nil
}
