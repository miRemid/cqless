package provider

import (
	"context"
	"net/url"

	"github.com/miRemid/cqless/pkg/types"
)

type ProviderPluginInterface interface {
	Init(*types.CQLessConfig) error // 初始化Plugin

	ValidNamespace(string) (bool, error) // 检查Namespace

	Deploy(ctx context.Context, req types.FunctionCreateRequest) (*types.Function, error)
	Remove(ctx context.Context, req types.FunctionRemoveRequest) (*types.Function, error)

	Resolve(ctx context.Context, functionName string) (url.URL, error)

	Close()
}
