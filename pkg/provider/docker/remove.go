package docker

import (
	"context"
	"sync"

	dtypes "github.com/docker/docker/api/types"
	"github.com/pkg/errors"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/pb"
)

func (p *DockerProvider) Remove(ctx context.Context, req *pb.DeleteFunctionRequest, cni *cninetwork.CNIManager) error {
	fns, err := p.getAllFunctionsByName(ctx, req.Name, cni)
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
		go func(fn *pb.Function) {
			defer wg.Done()
			if err := cni.DeleteCNINetwork(ctx, fn); err != nil {
				errChannel <- errors.WithMessagef(err, "删除CNI网络失败: ID=%s", fn.Id)
			}
			if err := p.cli.ContainerRemove(ctx, fn.Id, dtypes.ContainerRemoveOptions{
				Force: true,
			}); err != nil {
				errChannel <- errors.WithMessagef(err, "删除容器失败: ID=%s", fn.Id)
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
