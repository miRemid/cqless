package types

import (
	"errors"
	"time"
)

var (
	ErrFuncNotFound = errors.New("func not found")
	ErrNodeNotFound = errors.New("node not found")
)

const (
	SELECTOR_RANDOM = "RANDOM"

	STORAGE_LOCAL = "LOCAL"
	STORAGE_ETCD  = "ETCD"
)

type SelectorOption struct {
	Strategy string `yaml:"strategy" mapstructure:"strategy"`
}

type StorageOption struct {
	Strategy string `yaml:"strategy" mapstructure:"strategy"`

	RpcEndpoints []string      `yaml:"endpoints" mapstructure:"endpoints"`
	DialTimeout  time.Duration `yaml:"dial_timeout" mapstructure:"dial_timeout"`

	DBPath string `yaml:"db_path" mapstructure:"db_path"`
}

type ResolverOption struct {
	StorageOption  *StorageOption  `yaml:"storage" mapstructure:"storage"`
	SelectorOption *SelectorOption `yaml:"selector" mapstructure:"selector"`
}
