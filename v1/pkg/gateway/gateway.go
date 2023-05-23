package gateway

import (
	"context"

	"github.com/miRemid/cqless/v1/pkg/cninetwork"
	"github.com/miRemid/cqless/v1/pkg/gateway/types"
	"github.com/miRemid/cqless/v1/pkg/logger"
	"github.com/miRemid/cqless/v1/pkg/pb"
	"github.com/miRemid/cqless/v1/pkg/provider"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	// https://github.com/rfyiamcool/notes/blob/main/golang_net_http_optimize.md
	return nil
}

func (g *Gateway) CreateFunction(ctx context.Context, req *pb.CreateFunctionRequest) (*pb.Function, error) {
	return provider.Deploy(ctx, req, cninetwork.DefaultManager)
}

func (g *Gateway) UpdateFunction(context.Context, *pb.UpdateFunctionRequest) (*pb.Function, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateFunction not implemented")
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
