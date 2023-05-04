package provider

import (
	"context"
	"net/url"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"
)

type ProviderPluginInterface interface {
	Init(*types.GatewayConfig) error // 初始化Plugin

	ValidNamespace(string) (bool, error) // 检查Namespace

	Deploy(ctx context.Context, req types.FunctionCreateRequest, cni *cninetwork.CNIManager) (*types.Function, error)
	Remove(ctx context.Context, req types.FunctionRemoveRequest, cni *cninetwork.CNIManager) error
	Inspect(ctx context.Context, req types.FunctionInspectRequest, cni *cninetwork.CNIManager) ([]*types.Function, error)

	Resolve(ctx context.Context, functionName string, cni *cninetwork.CNIManager) (url.URL, error)

	Close()
}
