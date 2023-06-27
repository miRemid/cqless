package types

import "time"

type ProxyOption struct {
	Address string `yaml:"address" mapstructure:"address"`

	Timeout             time.Duration `yaml:"timeout" mapstructure:"timeout"`
	MaxIdleConns        int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host" mapstructure:"max_idle_conns_per_host"`

	NatsAddress string `yaml:"nats" mapstructure:"nats"`
}

const (
	ASYNC_REQUEST = "/async/request"
)
