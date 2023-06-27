package docker

import (
	"github.com/docker/docker/client"
	"github.com/miRemid/cqless/pkg/provider/types"
	dtypes "github.com/miRemid/cqless/pkg/types"
	"github.com/rs/zerolog"
)

type DockerProvider struct {
	cli *client.Client
	log zerolog.Logger
}

func newProvider() *DockerProvider {
	var p = new(DockerProvider)
	return p
}

func NewProvider() *DockerProvider {
	return newProvider()
}

func (p *DockerProvider) Init(config *types.ProviderOption, log zerolog.Logger) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	p.log = log
	if err != nil {
		return err
	}
	p.cli = cli
	return nil
}

func (p *DockerProvider) Close() error {
	return p.cli.Close()
}

func (p *DockerProvider) ValidNamespace(namespace string) (bool, error) {
	// TODO: 目前Docker仅支持默认namespace
	if namespace == dtypes.DEFAULT_FUNCTION_NAMESPACE {
		return true, nil
	}
	return false, nil
}
