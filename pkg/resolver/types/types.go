package types

import "errors"

var (
	ErrFuncNotFound = errors.New("func not found")
	ErrNodeNotFound = errors.New("node not found")
)

const (
	SELECTOR_RANDOM = "RANDOM"

	STORAGE_LOCAL = "LOCAL"
)

type SelectorOption struct {
	Strategy string `yaml:"strategy" mapstructure:"strategy"`
}

type StorageOption struct {
	Strategy string `yaml:"strategy" mapstructure:"strategy"`

	RpcEndpoint string `yaml:"endpoint" mapstructure:"endpoint"`

	DBPath string `yaml:"db_path" mapstructure:"db_path"`
}

type ResolverOption struct {
	StorageOption  *StorageOption  `yaml:"storage" mapstructure:"storage"`
	SelectorOption *SelectorOption `yaml:"selector" mapstructure:"selector"`
}
