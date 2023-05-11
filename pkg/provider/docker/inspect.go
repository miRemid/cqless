package docker

import (
	"context"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"
)

func (p *DockerProvider) Inspect(ctx context.Context, req types.FunctionInspectRequest, cni *cninetwork.CNIManager) ([]*types.Function, error) {
	log.Debug().Msg("Inspect function")
	var functions []*types.Function
	var err error
	if req == (types.FunctionInspectRequest{}) {
		functions, err = p.getAllFunctions(ctx, cni)
	} else {
		functions, err = p.getAllFunctionsByName(ctx, req.FunctionName, cni)
	}
	if err != nil {
		return nil, err
	}
	return functions, nil
}
