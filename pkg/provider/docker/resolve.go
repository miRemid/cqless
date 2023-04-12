package docker

import (
	"context"
	"fmt"
	"net/url"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/provider"
)

func (p *DockerProvider) Resolve(ctx context.Context, functionName string, cni *cninetwork.CNIManager) (url.URL, error) {
	fns, err := p.getAllFunctionsByName(ctx, functionName, cni)
	if err != nil {
		return url.URL{}, err
	}
	fn := fns[0]
	urlStr := fmt.Sprintf("http://%s:%s", fn.IPAddress, provider.WatchdogPort)
	urlRes, err := url.Parse(urlStr)
	return *urlRes, err
}
