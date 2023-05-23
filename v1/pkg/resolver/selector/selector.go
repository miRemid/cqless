package selector

import (
	"context"
	"strings"

	"github.com/miRemid/cqless/v1/pkg/resolver/types"
)

// 负载均衡
type SelectorInterface interface {
	Next(context.Context) (*types.Node, error)
	Add(context.Context, ...*types.Node) error
	Del(context.Context, *types.Node) error
}

func NewSelector(opt *types.SelectorOption) SelectorInterface {
	switch strings.ToUpper(opt.Strategy) {
	case types.SELECTOR_RANDOM:
		return newRandomSelector(opt)
	default:
		return newRandomSelector(opt)
	}
}
