package docker

import (
	"context"

	"github.com/miRemid/cqless/v1/pkg/cninetwork"
	"github.com/miRemid/cqless/v1/pkg/pb"
)

func (p *DockerProvider) Inspect(ctx context.Context, req *pb.GetFunctionRequest, cni *cninetwork.CNIManager) ([]*pb.Function, error) {
	p.log.Debug().Msg("Inspect function")
	var functions []*pb.Function
	var err error
	if req.Name == "" {
		functions, err = p.getAllFunctions(ctx, cni)
	} else {
		functions, err = p.getAllFunctionsByName(ctx, req.Name, cni)
	}
	if err != nil {
		return nil, err
	}
	return functions, nil
}
