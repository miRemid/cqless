package storage

import (
	"context"
	"strings"

	local "github.com/miRemid/cqless/pkg/resolver/storage/local"
	"github.com/miRemid/cqless/pkg/resolver/types"
	"github.com/rs/zerolog"
)

type StorageInterface interface {
	Get(ctx context.Context, funcName string) ([]*types.Node, error)
	Register(ctx context.Context, funcName string, node *types.Node) error
	UnRegister(ctx context.Context, funcName string, node *types.Node) error
	UnRegisterFunc(ctx context.Context, funcName string) error

	Close() error
	Init() error
}

func NewStorage(opt *types.StorageOption, log zerolog.Logger) StorageInterface {
	storage := strings.ToUpper(opt.Strategy)
	sublog := log.With().Str("storage", storage).Logger()
	switch storage {
	case types.STORAGE_LOCAL:
		return local.New(opt, sublog)
	default:
		return local.New(opt, sublog)
	}
}
