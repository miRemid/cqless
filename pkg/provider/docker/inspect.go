package docker

import (
	"context"

	dtypes "github.com/docker/docker/api/types"
)

func (p *DockerProvider) Inspect(ctx context.Context, id string) (dtypes.ContainerJSON, error) {
	return p.cli.ContainerInspect(ctx, id)
}
