package docker

import (
	"context"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/rs/zerolog/log"
)

func (p *DockerProvider) Inspect(ctx context.Context, req types.FunctionGetRequest, cni *cninetwork.CNIManager) ([]*types.Function, error) {
	log.Debug().Msg("Inspect function")
	var functions []*types.Function
	var err error
	if req == (types.FunctionGetRequest{}) {
		functions, err = p.getAllFunctions(ctx, cni)
	} else {
		functions, err = p.getAllFunctionsByName(ctx, req.FunctionName, cni)
	}
	if err != nil {
		return nil, err
	}
	return functions, nil
}
