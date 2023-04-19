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
	if len(fns) == 0 {
		return url.URL{}, fmt.Errorf("未发现和 '%s' 函数相关容器", functionName)
	}
	fn := fns[0]
	urlStr := fmt.Sprintf("http://%s:%s", fn.IPAddress, provider.WatchdogPort)
	urlRes, err := url.Parse(urlStr)
	return *urlRes, err
}
