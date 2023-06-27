package types

import (
	"fmt"

	"github.com/miRemid/cqless/v1/pkg/pb"
	"github.com/miRemid/cqless/v1/pkg/resolver/types"
)

func BuildEnv(function *pb.Function) []string {
	var env = make([]string, 0)
	for k, v := range function.Envs {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

func BuildNode(function *pb.Function) *types.Node {
	return types.NewNode(function.Scheme, function.IpAddress+":"+function.WatchDogPort, function.Name, function.Metadata)
}

type FunctionInvokeRequest struct {
	FunctionName string `form:"name" json:"name"`
}
