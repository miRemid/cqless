package selector

import (
	"context"
	"math/rand"
	"sort"
	"time"

	"github.com/miRemid/cqless/v1/pkg/resolver/types"
)

type randomSelector struct {
	nodes []*types.Node
	opt   *types.SelectorOption
}

func newRandomSelector(opt *types.SelectorOption) SelectorInterface {
	return &randomSelector{
		nodes: make([]*types.Node, 0),
		opt:   opt,
	}
}

func (s *randomSelector) Add(ctx context.Context, node ...*types.Node) error {
	s.nodes = append(s.nodes, node...)
	return nil
}

func (s *randomSelector) Next(ctx context.Context) (*types.Node, error) {
	rander := rand.New(rand.NewSource(time.Now().UnixMicro()))
	idx := rander.Perm(len(s.nodes))
	return s.nodes[idx[0]], nil
}

func (s *randomSelector) Del(ctx context.Context, node *types.Node) error {
	idx := sort.Search(len(s.nodes), func(i int) bool {
		return s.nodes[i] == node
	})
	if idx == len(s.nodes) {
		return types.ErrNodeNotFound
	}
	s.nodes = append(s.nodes[:idx], s.nodes[idx+1:]...)
	return nil
}

func (s *randomSelector) Close() error {
	return nil
}
