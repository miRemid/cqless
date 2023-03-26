package docker

import (
	"context"
	"net/url"
)

func (p *DockerProvider) Resolve(ctx context.Context, functionName string) (url.URL, error) {

	return url.URL{}, nil
}
