package docker

import (
	"context"
	"sync"

	dtypes "github.com/docker/docker/api/types"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/pkg/errors"
)

func (p *DockerProvider) Remove(ctx context.Context, req types.FunctionRemoveRequest, cni *cninetwork.CNIManager) error {
	fns, err := p.getAllFunctionsByName(ctx, req.FunctionName, cni)
	if err != nil {
		return errors.WithMessage(err, "获取容器列表失败")
	}
	if len(fns) == 0 {
		return errors.Errorf("未找到相关容器")
	}
	wg := sync.WaitGroup{}
	errChannel := make(chan error, len(fns))
	for _, fn := range fns {
		wg.Add(1)
		go func(fn *types.Function) {
			defer wg.Done()
			if err := cni.DeleteCNINetwork(ctx, fn); err != nil {
				errChannel <- errors.WithMessagef(err, "删除CNI网络失败: ID=%s", fn.ID)
			}
			if err := p.cli.ContainerRemove(ctx, fn.ID, dtypes.ContainerRemoveOptions{
				Force: true,
			}); err != nil {
				errChannel <- errors.WithMessagef(err, "删除容器失败: ID=%s", fn.ID)
			}
		}(fn)
	}
	wg.Wait()
	close(errChannel)
	err = nil
	for e := range errChannel {
		err = errors.Wrap(err, e.Error())
	}
	return err
}
