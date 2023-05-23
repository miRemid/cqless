package types

import "time"

type GatewayOption struct {
	Address      string        `yaml:"address" mapstructure:"address"`
	HTTPAddress  string        `yaml:"http_address" mapstructure:"http_address"`
	ReadTimeout  time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`

	EnableRateLimit bool             `yaml:"enable_rate_limit" mapstructure:"enable_rate_limit"`
	RateLimit       *RateLimitOption `yaml:"rate_limit" mapstructure:"rate_limit"`

	EnablePprof bool `yaml:"pprof" mapstructure:"pprof"`
}

type RateLimitOption struct {
	Limit float64 `yaml:"limit" mapstructure:"limit"`
	Burst int     `yaml:"burst" mapstructure:"burst"`
}
