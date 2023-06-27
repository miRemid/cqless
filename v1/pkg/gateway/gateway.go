package gateway

import (
	"context"

	"github.com/miRemid/cqless/v1/pkg/cninetwork"
	"github.com/miRemid/cqless/v1/pkg/gateway/types"
	"github.com/miRemid/cqless/v1/pkg/logger"
	"github.com/miRemid/cqless/v1/pkg/pb"
	"github.com/miRemid/cqless/v1/pkg/provider"
	"github.com/miRemid/cqless/v1/pkg/resolver"
	rtypes "github.com/miRemid/cqless/v1/pkg/resolver/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	defaultGateway *Gateway
)

func init() {
	defaultGateway = new(Gateway)
}

func Init(config *types.GatewayOption) error {
	return defaultGateway.Init(config)
}

type Gateway struct {
	pb.UnimplementedFunctionServiceServer
	log zerolog.Logger
}

func (gate *Gateway) Init(config *types.GatewayOption) error {
	gate.log = log.Hook(logger.ModuleHook("gateway"))
	return nil
}

func (g *Gateway) CreateFunction(ctx context.Context, req *pb.CreateFunctionRequest) (*pb.Function, error) {
	function, err := provider.Deploy(ctx, req, cninetwork.DefaultManager)
	if err != nil {
		return nil, err
	}
	if err := resolver.Register(ctx, function.Name, rtypes.NewNode(
		function.Scheme,
		function.IpAddress+":"+function.WatchDogPort,
		function.Name, function.Metadata,
	)); err != nil {
		defer provider.Remove(context.Background(), &pb.DeleteFunctionRequest{
			Name: function.Name,
		}, cninetwork.DefaultManager)
		return nil, err
	}
	return function, nil
}

func (g *Gateway) ListFunctions(ctx context.Context, req *pb.ListFunctionsRequest) (*pb.ListFunctionsResponse, error) {
	fns, err := provider.Inspect(ctx, &pb.GetFunctionRequest{}, cninetwork.DefaultManager)
	if err != nil {
		return nil, err
	}
	return &pb.ListFunctionsResponse{
		Functions: fns,
	}, nil
}

func (g *Gateway) GetFunction(ctx context.Context, req *pb.GetFunctionRequest) (*pb.ListFunctionsResponse, error) {
	fns, err := provider.Inspect(ctx, req, cninetwork.DefaultManager)
	if err != nil {
		return nil, err
	}
	return &pb.ListFunctionsResponse{
		Functions: fns,
	}, nil
}

func (g *Gateway) DeleteFunction(ctx context.Context, req *pb.DeleteFunctionRequest) (*emptypb.Empty, error) {
	err := provider.Remove(ctx, req, cninetwork.DefaultManager)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func New() *Gateway {
	return &Gateway{}
}
