package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/spf13/viper"
)

const (
	DefaultInterval = 30 * time.Second
	DefaultTimeout  = 5 * time.Second
)

type Config struct {
	Interval time.Duration `mapstructure:"interval"`
	Timeout  time.Duration `mapstructure:"timeout"`
	Targets  []Target      `mapstructure:"targets"`
}

type Target struct {
	Name     string        `mapstructure:"name"`
	URL      string        `mapstructure:"url"`
	Type     string        `mapstructure:"type"`
	Interval time.Duration `mapstructure:"interval"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

func (cfg *Config) validate() error {
	if cfg.Interval <= 0 {
		slog.Debug("invalid global interval")
		cfg.Interval = DefaultInterval
	}
	if cfg.Timeout <= 0 {
		slog.Debug("invalid global timeout")
		cfg.Timeout = DefaultTimeout
	}

	for i := range cfg.Targets {
		trg := &cfg.Targets[i]
		parsed, err := url.Parse(trg.URL)
		if err != nil {
			return err
		}
		if !isValidType(trg.Type) {
			return fmt.Errorf("unsupported type")
		}
		if trg.Name == "" {
			trg.Name = parsed.Host
		}
		if trg.Interval <= 0 {
			trg.Interval = cfg.Interval
		}
		if trg.Timeout <= 0 {
			trg.Timeout = cfg.Timeout
		}

	}

	return nil
}

func isValidType(targetType string) bool {
	switch targetType {
	case "http", "https", "tcp", "grpc":
		return true
	default:
		return false
	}
}

func LoadConfig() (*Config, error) {
	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
