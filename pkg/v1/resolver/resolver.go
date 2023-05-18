package resolver

import (
	"context"
	"sync"

	"github.com/miRemid/cqless/pkg/v1/logger"
	"github.com/miRemid/cqless/pkg/v1/resolver/selector"
	"github.com/miRemid/cqless/pkg/v1/resolver/storage"
	"github.com/miRemid/cqless/pkg/v1/resolver/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Resolver struct {
	storage storage.StorageInterface
	cache   map[string]selector.SelectorInterface
	rwmutex sync.RWMutex
	log     zerolog.Logger

	opt *types.ResolverOption
}

var defaultResolver *Resolver

func init() {
	defaultResolver = new(Resolver)
}

func Init(opt *types.ResolverOption) error {
	return defaultResolver.Init(opt)
}

func (r *Resolver) Init(opt *types.ResolverOption) error {
	r.opt = opt
	r.log = log.Hook(logger.ModuleHook("resolver"))
	r.storage = storage.NewStorage(opt.StorageOption, r.log)
	r.cache = make(map[string]selector.SelectorInterface)
	return r.storage.Init()
}

func Next(ctx context.Context, funcName string) (*types.Node, error) {
	return defaultResolver.Next(ctx, funcName)
}

func (r *Resolver) Next(ctx context.Context, funcName string) (*types.Node, error) {
	evt := r.log.With().Str("action", "next").Str("funcName", funcName).Logger()
	r.rwmutex.RLock()
	s, ok := r.cache[funcName]
	if ok {
		node, err := s.Next(ctx)
		if err == nil {
			evt.Debug().Str("node", node.String()).Msg("found func node")
			return node, nil
		}
	}
	r.rwmutex.RUnlock()
	nodes, err := r.storage.Get(ctx, funcName)
	if err != nil {
		evt.Debug().Err(err).Msg("func node not found")
		return nil, types.ErrNodeNotFound
	}
	r.rwmutex.Lock()
	if !ok {
		evt.Debug().Str("selector", funcName).Msg("selector not found, create")
		r.cache[funcName] = selector.NewSelector(r.opt.SelectorOption)
	}
	r.cache[funcName].Add(ctx, nodes...)
	r.rwmutex.Unlock()
	if len(nodes) == 0 {
		return nil, types.ErrNodeNotFound
	}
	evt.Debug().Str("node", nodes[0].String()).Msg("found func node")
	return nodes[0], nil
}

func Register(ctx context.Context, funcName string, node *types.Node) error {
	return defaultResolver.Register(ctx, funcName, node)
}

func (r *Resolver) Register(ctx context.Context, funcName string, node *types.Node) error {
	if err := r.storage.Register(ctx, funcName, node); err != nil {
		return err
	}
	_, ok := r.cache[funcName]
	if !ok {
		r.cache[funcName] = selector.NewSelector(r.opt.SelectorOption)
	}
	if err := r.cache[funcName].Add(ctx, node); err != nil {
		r.storage.UnRegister(ctx, funcName, node)
		return err
	}
	return nil
}

func UnRegister(ctx context.Context, funcName string, node *types.Node) error {
	return defaultResolver.UnRegister(ctx, funcName, node)
}

func (r *Resolver) UnRegister(ctx context.Context, funcName string, node *types.Node) error {
	if s, ok := r.cache[funcName]; ok {
		s.Del(ctx, node)
	}
	if err := r.storage.UnRegister(ctx, funcName, node); err != nil {
		return err
	}
	return nil
}

func UnRegisterFunc(ctx context.Context, funcName string) error {
	return defaultResolver.UnRegisterFunc(ctx, funcName)
}

func (r *Resolver) UnRegisterFunc(ctx context.Context, funcName string) error {
	delete(r.cache, funcName)
	if err := r.storage.UnRegisterFunc(ctx, funcName); err != nil {
		return err
	}
	return nil
}
