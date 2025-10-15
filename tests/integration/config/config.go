package config

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Lambda LambdaConfig `mapstructure:"lambda"`
}

type ServerConfig struct {
	Host     string `mapstructure:"host"`
	Protocol string `mapstructure:"protocol"`
	Port     int    `mapstructure:"port"`
}

type LambdaConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

var (
	loadOnce sync.Once
	onceCfg  *Config
	onceErr  error
)

func (s *ServerConfig) URL() string {
	protocol := s.Protocol
	if protocol == "" {
		protocol = "http"
	}
	// Don't include port if it's the default for the protocol
	if (protocol == "http" && s.Port == 80) || (protocol == "https" && s.Port == 443) {
		return fmt.Sprintf("%s://%s", protocol, s.Host)
	}
	return fmt.Sprintf("%s://%s:%d", protocol, s.Host, s.Port)
}

func (l *LambdaConfig) InvocationURL() string {
	return fmt.Sprintf("http://%s:%d/2015-03-31/functions/function/invocations", l.Host, l.Port)
}

func loadConfig() (*Config, error) {
	k := koanf.New(".")
	if err := k.Load(file.Provider(filepath.Join(configDir(), "default.yml")), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("error reading default config: %w", err)
	}
	// Attempt local merge
	_ = k.Load(file.Provider(filepath.Join(configDir(), "local.yml")), yaml.Parser())
	// Attempt private merge
	_ = k.Load(file.Provider(filepath.Join(configDir(), "private.yml")), yaml.Parser())

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}
	return &cfg, nil
}

func configDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Could not get current file info")
	}

	dir := filepath.Dir(filename)
	return filepath.Clean(dir)
}

// LoadConfig loads the config once and returns it.
// Reusing within the same test suite will only load once.
func LoadConfig(t *testing.T) *Config {
	t.Helper()
	loadOnce.Do(func() {
		onceCfg, onceErr = loadConfig()
	})
	if onceErr != nil {
		t.Fatalf("failed to load config: %v", onceErr)
	}
	return onceCfg
}
