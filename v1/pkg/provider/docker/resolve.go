package docker

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/miRemid/cqless/v1/pkg/cninetwork"
)

func (p *DockerProvider) Resolve(ctx context.Context, functionName string, cni *cninetwork.CNIManager) (url.URL, error) {
	fns, err := p.getAllFunctionsByName(ctx, functionName, cni)
	if err != nil {
		return url.URL{}, err
	}
	if len(fns) == 0 {
		return url.URL{}, fmt.Errorf("未发现和 '%s' 函数相关容器", functionName)
	}
	// TODO: 负载均衡，目前随机选取
	newRand := rand.New(rand.NewSource(time.Now().Unix()))
	idx := newRand.Intn(len(fns))
	fn := fns[idx]
	var urlStr string
	if fn.WatchDogPort != "" {
		urlStr = fmt.Sprintf("http://%s:%s", fn.IpAddress, fn.WatchDogPort)
	} else {
		urlStr = fmt.Sprintf("http://%s", fn.IpAddress)
	}
	urlRes, err := url.Parse(urlStr)
	return *urlRes, err
}
