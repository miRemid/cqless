package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"

	"github.com/miRemid/cqless/pkg/v1/resolver/types"
	"github.com/nutsdb/nutsdb"
	"github.com/rs/zerolog"
)

/*

bucket: cqless

SetKey: {function}, Value {scheme}://{host}

ValueKey: {function}/{scheme}://{hostpath}: types.Node

*/

const (
	bucket = "cqless"
)

type localStorage struct {
	log zerolog.Logger
	opt *types.StorageOption
	db  *nutsdb.DB
}

func New(opt *types.StorageOption, log zerolog.Logger) *localStorage {
	return &localStorage{
		log: log,
		opt: opt,
	}
}

func (r *localStorage) Init() error {
	// open db
	evt := r.log.With().Str("action", "init").Logger()
	evt.Debug().Str("db_path", r.opt.DBPath).Msg("open db")
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(path.Join(r.opt.DBPath, "db")))
	if err != nil {
		evt.Debug().Err(err).Msg("init failed")
		return err
	}
	r.db = db
	return nil
}

func (r *localStorage) Close() error {
	evt := r.log.With().Str("action", "close").Logger()
	if err := r.db.Close(); err != nil {
		evt.Debug().Err(err).Msg("close db failed")
		return err
	}
	return nil
}

func (r *localStorage) Get(ctx context.Context, funcName string) ([]*types.Node, error) {
	evt := r.log.With().Str("action", "get").Str("func", funcName).Logger()
	var nodes = make([]*types.Node, 0)
	if err := r.db.View(func(tx *nutsdb.Tx) error {
		data, err := tx.SMembers(bucket, []byte(funcName))
		if err != nil {
			return err
		}
		// d: {scheme}://{host}
		for _, d := range data {
			// valueKey: {funcName}/{scheme}://{host}
			valueKey := fmt.Sprintf("%s/%s", funcName, string(d))
			value, err := tx.Get(bucket, []byte(valueKey))
			if err != nil {
				evt.Debug().Err(err).Str("key", string(value.Key)).Msg("get node info failed, continue...")
				continue
			}
			var node = new(types.Node)
			var buffer = bytes.NewBuffer(value.Value)
			if err := json.NewDecoder(buffer).Decode(node); err != nil {
				evt.Debug().Err(err).Msg("decode node failed, continue...")
				continue
			}
			nodes = append(nodes, node)
		}
		evt.Debug().Int("nodes-num", len(nodes)).Msg("get nodes")
		return nil
	}); err != nil {
		evt.Debug().Err(err).Msg("get faield")
		return nil, err
	}
	return nodes, nil
}

func (r *localStorage) Register(ctx context.Context, funcName string, node *types.Node) error {
	evt := r.log.With().Str("action", "register").Str("func", funcName).Logger()
	if err := r.db.Update(func(tx *nutsdb.Tx) error {
		var buffer bytes.Buffer
		if err := json.NewEncoder(&buffer).Encode(node); err != nil {
			return err
		}
		valueKey := node.GetValueKey()
		evt.Debug().Str("valueKey", valueKey).Msg("put valueKey")
		if err := tx.Put(bucket, []byte(valueKey), buffer.Bytes(), 0); err != nil {
			evt.Debug().Err(err).Str("vetKey", valueKey).Msg("put valueKey failed")
			return err
		}
		evt.Debug().Str("setKey", node.String()).Msg("put setKey")
		if err := tx.SAdd(bucket, []byte(funcName), []byte(node.String())); err != nil {
			evt.Debug().Err(err).Str("setKey", node.String()).Msg("put setKey failed")
			return err
		}
		return nil
	}); err != nil {
		evt.Debug().Err(err).Msg("register failed")
		return err
	}
	return nil
}

func (r *localStorage) UnRegister(ctx context.Context, funcName string, node *types.Node) error {
	evt := r.log.With().Str("action", "unregister").Str("func", funcName).Logger()
	if err := r.db.Update(func(tx *nutsdb.Tx) error {
		valueKey := node.GetValueKey()
		evt.Debug().Str("valueKey", node.String()).Msg("delete valueKey")
		if err := tx.Delete(bucket, []byte(valueKey)); err != nil {
			evt.Debug().Err(err).Str("valueKey", valueKey).Msg("delete valueKey failed")
			return err
		}

		evt.Debug().Str("setKey", node.String()).Msg("delete SetKey")
		if err := tx.SRem(bucket, []byte(funcName), []byte(node.String())); err != nil {
			evt.Debug().Err(err).Str("setKey", node.String()).Msg("delete setKey failed")
			return err
		}
		return nil
	}); err != nil {
		evt.Debug().Err(err).Msg("unregister failed")
		return err
	}
	return nil
}

func (r *localStorage) UnRegisterFunc(ctx context.Context, funcName string) error {
	evt := r.log.With().Str("action", "unregister-func").Str("func", funcName).Logger()
	if err := r.db.Update(func(tx *nutsdb.Tx) error {
		entries, _, err := tx.PrefixScan(bucket, []byte(funcName), 0, -1)
		evt.Debug().Msg("scan function entries")
		if err != nil {
			evt.Debug().Err(err).Msg("scan function entries failed")
			return err
		}
		evt.Debug().Int("num", len(entries)).Msg("delete function entries")
		for _, entry := range entries {
			if err := tx.Delete(bucket, entry.Key); err != nil {
				evt.Debug().Err(err).Str("key", string(entry.Key)).Msg("delete function entry failed, skip")
				continue
			}
		}
		return nil
	}); err != nil {
		evt.Debug().Err(err).Msg("unregister failed")
		return err
	}
	return nil
}
