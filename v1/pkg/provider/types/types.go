package types

const (
	PROVIDER_DOCKER = "DOCKER"
)

type ProviderOption struct {
	Strategy string `yaml:"strategy" mapstructure:"strategy"`
}
