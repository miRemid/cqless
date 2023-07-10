package etcd

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/miRemid/cqless/pkg/resolver/types"
)

const (
	bucket   = "cqless"
	lockName = "/lock/cqless"
)

type etcdStorage struct {
	log zerolog.Logger
	opt *types.StorageOption
	cli *clientv3.Client
}

func New(opt *types.StorageOption, log zerolog.Logger) *etcdStorage {
	return &etcdStorage{
		log: log,
		opt: opt,
	}
}

func (s *etcdStorage) getLock(ctx context.Context, funcName string) (*concurrency.Mutex, error) {
	session, err := concurrency.NewSession(s.cli, concurrency.WithTTL(5))
	if err != nil {
		return nil, err
	}
	lock := concurrency.NewMutex(session, lockName+"-"+funcName)
	if err := lock.Lock(context.TODO()); err != nil {
		return nil, err
	}
	return lock, nil
}

func (s *etcdStorage) Init() error {
	evt := s.log.With().Str("action", "init").Logger()
	evt.Debug().Strs("etcd-endpoint", s.opt.RpcEndpoints).Msg("connect to remote etcd server")
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   s.opt.RpcEndpoints,
		DialTimeout: s.opt.DialTimeout,
	})
	if err != nil {
		return err
	}
	s.cli = cli
	return nil
}

func (s *etcdStorage) Close() error {
	evt := s.log.With().Str("action", "close").Logger()
	if err := s.cli.Close(); err != nil {
		evt.Debug().Err(err).Msg("close connection from etcd failed")
		return err
	}
	return nil
}

func (s *etcdStorage) Get(ctx context.Context, funcName string) ([]*types.Node, error) {
	var nodes = make([]*types.Node, 0)
	lock, err := s.getLock(ctx, funcName)
	if err != nil {
		return nil, err
	}
	defer lock.Unlock(ctx)
	// get all nodes
	// prefix: funcName
	resp, err := s.cli.Get(ctx, funcName, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	for _, kv := range resp.Kvs {
		var node = new(types.Node)
		json.Unmarshal(kv.Value, node)
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (s *etcdStorage) Register(ctx context.Context, funcName string, node *types.Node) error {
	lock, err := s.getLock(ctx, funcName)
	if err != nil {
		return err
	}
	defer lock.Unlock(ctx)
	_, err = s.cli.Put(ctx, node.GetValueKey(), string(node.Bytes()))
	if err != nil {
		return err
	}
	return nil
}

func (s *etcdStorage) UnRegister(ctx context.Context, funcName string, node *types.Node) error {
	lock, err := s.getLock(ctx, funcName)
	if err != nil {
		return err
	}
	defer lock.Unlock(ctx)
	_, err = s.cli.Delete(ctx, node.GetValueKey())
	if err != nil {
		return err
	}
	return nil
}

func (s *etcdStorage) UnRegisterFunc(ctx context.Context, funcName string) error {
	lock, err := s.getLock(ctx, funcName)
	if err != nil {
		return err
	}
	defer lock.Unlock(ctx)
	_, err = s.cli.Delete(ctx, funcName, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	return nil
}
