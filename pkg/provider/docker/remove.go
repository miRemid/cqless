package docker

import (
	"context"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"
)

func (p *DockerProvider) Remove(ctx context.Context, req types.FunctionRemoveRequest, cni *cninetwork.CNIManager, namespace string) error {
	return nil
}
