package docker

import (
	"context"

	dtypes "github.com/docker/docker/api/types"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/pkg/errors"
)

func (p *DockerProvider) Remove(ctx context.Context, req types.FunctionRemoveRequest, cni *cninetwork.CNIManager) (*types.Function, error) {
	fns, err := p.getAllFunctionsByName(ctx, req.FunctionName, cni)
	if err != nil {
		return nil, errors.WithMessage(err, "get function container error")
	}
	// TODO: 目前只能删除第一个容器，暂不支持全部删除
	fn := fns[0]
	if err := cni.DeleteCNINetwork(ctx, fn); err != nil {
		return nil, errors.WithMessage(err, "delete cni network error")
	}
	if err = p.cli.ContainerRemove(ctx, fn.ID, dtypes.ContainerRemoveOptions{
		Force: true,
	}); err != nil {
		return nil, errors.WithMessage(err, "delete container error")
	}
	return fn, nil
}
