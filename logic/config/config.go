package config

import "time"

const (
	DefaultInterval = 30 * time.Second
	DefaultTimeout  = 5 * time.Second
)

type Config struct {
	Interval string   `mapstructure:"interval"`
	Timeout  string   `mapstructure:"timeout"`
	Targets  []Target `mapstructure:"targets"`
}

type Target struct {
	Name     string `mapstructure:"name"`
	URL      string `mapstructure:"url"`
	Type     string `mapstructure:"type"`
	Interval string `mapstructure:"interval"`
	Timeout  string `mapstructure:"timeout"`
}
