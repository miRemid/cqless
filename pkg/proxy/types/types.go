package types

import "time"

type ProxyOption struct {
	Port int `yaml:"port" mapstructure:"port"`

	Timeout             time.Duration `yaml:"timeout" mapstructure:"timeout"`
	MaxIdleConns        int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host" mapstructure:"max_idle_conns_per_host"`
}
