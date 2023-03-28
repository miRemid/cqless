package docker

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/miRemid/cqless/pkg/provider"
	"github.com/miRemid/cqless/pkg/types"
	zerolog "github.com/rs/zerolog/log"
)

var log = zerolog.With().Str("provider", "docker").Logger()

type DockerProvider struct {
	cli   *client.Client
	store *provider.FakeLabeller
}

func newProvider() *DockerProvider {
	var p = new(DockerProvider)
	p.store = new(provider.FakeLabeller)
	return p
}

func NewProvider() provider.ProviderPluginInterface {
	return newProvider()
}

func (p *DockerProvider) Init(config *types.CQLessConfig) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	p.cli = cli
	return nil
}

func (p *DockerProvider) Close() {
	p.cli.Close()
}

func (p *DockerProvider) ValidNamespace(namespace string) (bool, error) {
	if namespace == types.DEFAULT_FUNCTION_NAMESPACE {
		return true, nil
	}
	labels, err := p.store.Labels(context.Background(), namespace)
	if err != nil {
		return false, err
	}
	if v, found := labels[types.NAMESPACE_LABEL]; found && v == "true" {
		return true, nil
	}
	return false, nil
}
