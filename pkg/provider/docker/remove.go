package docker

import (
	"context"

	dtypes "github.com/docker/docker/api/types"
	"github.com/miRemid/cqless/pkg/types"
)

func (p *DockerProvider) Remove(ctx context.Context, req types.FunctionRemoveRequest) (*types.Function, error) {
	fn, err := p.getFunction(ctx, req.FunctionName)
	if err != nil {
		return nil, err
	}
	err = p.cli.ContainerRemove(ctx, fn.ID, dtypes.ContainerRemoveOptions{
		Force: true,
	})
	return fn, err
}
