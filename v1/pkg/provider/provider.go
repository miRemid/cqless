package provider

import (
	"context"
	"net/url"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/miRemid/cqless/v1/pkg/cninetwork"
	"github.com/miRemid/cqless/v1/pkg/logger"
	"github.com/miRemid/cqless/v1/pkg/pb"
	"github.com/miRemid/cqless/v1/pkg/provider/docker"
	"github.com/miRemid/cqless/v1/pkg/provider/types"
)

var defaultProvider *Provider

func init() {
	defaultProvider = new(Provider)
}

type ProviderPluginInterface interface {
	Init(*types.ProviderOption, zerolog.Logger) error // 初始化Plugin

	Deploy(ctx context.Context, req *pb.CreateFunctionRequest, cni *cninetwork.CNIManager) (*pb.Function, error)
	Remove(ctx context.Context, req *pb.DeleteFunctionRequest, cni *cninetwork.CNIManager) error
	Inspect(ctx context.Context, req *pb.GetFunctionRequest, cni *cninetwork.CNIManager) ([]*pb.Function, error)
	Resolve(ctx context.Context, functionName string, cni *cninetwork.CNIManager) (url.URL, error)

	Close() error
}

type Provider struct {
	plugin ProviderPluginInterface
	log    zerolog.Logger

	opt *types.ProviderOption
}

func Init(opt *types.ProviderOption) error {
	return defaultProvider.Init(opt)
}

func (p *Provider) Init(opt *types.ProviderOption) error {
	p.log = log.Hook(logger.ModuleHook("provider"))
	p.opt = opt
	plugin := strings.ToUpper(opt.Strategy)
	pluginLog := p.log.With().Str("plugin", plugin).Logger()
	switch plugin {
	case types.PROVIDER_DOCKER:
		p.plugin = docker.NewProvider()
	default:
		p.plugin = docker.NewProvider()
	}
	return p.plugin.Init(p.opt, pluginLog)
}

func (p *Provider) Deploy(ctx context.Context, req *pb.CreateFunctionRequest, cni *cninetwork.CNIManager) (*pb.Function, error) {
	return p.plugin.Deploy(ctx, req, cni)
}
func (p *Provider) Remove(ctx context.Context, req *pb.DeleteFunctionRequest, cni *cninetwork.CNIManager) error {
	return p.plugin.Remove(ctx, req, cni)
}
func (p *Provider) Inspect(ctx context.Context, req *pb.GetFunctionRequest, cni *cninetwork.CNIManager) ([]*pb.Function, error) {
	return p.plugin.Inspect(ctx, req, cni)
}
func (p *Provider) Resolve(ctx context.Context, functionName string, cni *cninetwork.CNIManager) (url.URL, error) {
	return p.plugin.Resolve(ctx, functionName, cni)
}
func (p *Provider) Close() error {
	return p.plugin.Close()
}

func Deploy(ctx context.Context, req *pb.CreateFunctionRequest, cni *cninetwork.CNIManager) (*pb.Function, error) {
	return defaultProvider.Deploy(ctx, req, cni)
}
func Remove(ctx context.Context, req *pb.DeleteFunctionRequest, cni *cninetwork.CNIManager) error {
	return defaultProvider.Remove(ctx, req, cni)
}
func Inspect(ctx context.Context, req *pb.GetFunctionRequest, cni *cninetwork.CNIManager) ([]*pb.Function, error) {
	return defaultProvider.Inspect(ctx, req, cni)
}
func Resolve(ctx context.Context, functionName string, cni *cninetwork.CNIManager) (url.URL, error) {
	return defaultProvider.Resolve(ctx, functionName, cni)
}
func Close() error {
	return defaultProvider.Close()
}
