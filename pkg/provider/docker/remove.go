package docker

import (
	"context"

	dtypes "github.com/docker/docker/api/types"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"
)

func (p *DockerProvider) Remove(ctx context.Context, req types.FunctionRemoveRequest, cni *cninetwork.CNIManager) (*types.Function, error) {
	fn, err := p.getFunction(ctx, req.FunctionName, cni)
	if err != nil {
		return nil, err
	}
	if err := cni.DeleteCNINetwork(ctx, fn); err != nil {
		return nil, err
	}
	if err = p.cli.ContainerRemove(ctx, fn.ID, dtypes.ContainerRemoveOptions{
		Force: true,
	}); err != nil {
		return nil, err
	}
	return fn, err
}
