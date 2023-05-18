package etcd_test

import (
	"context"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"gotest.tools/v3/assert"

	"github.com/miRemid/cqless/pkg/v1/resolver/types"
)

var (
	opt = &types.StorageOption{
		Strategy:     types.STORAGE_ETCD,
		RpcEndpoints: []string{"127.0.0.1:12379"},
		DialTimeout:  2 * time.Second,
	}
	funcName = "test"
	node1    = types.NewNode("http", "127.0.0.1:8080", funcName, nil)
	node2    = types.NewNode("wss", "127.0.0.1:8081", funcName, nil)

	client *clientv3.Client
)

func init() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   opt.RpcEndpoints,
		DialTimeout: opt.DialTimeout,
	})
	if err != nil {
		panic(err)
	}
	client = cli
}

func Test_etcd(t *testing.T) {
	// 1. put
	// Key: /funcName/scheme/Host
	// Value: types.Node
	_, err := client.Put(context.TODO(), node1.GetValueKey(), string(node1.Bytes()))
	assert.NilError(t, err)
	_, err = client.Put(context.TODO(), node2.GetValueKey(), string(node2.Bytes()))
	assert.NilError(t, err)

	res, err := client.Get(context.TODO(), funcName, clientv3.WithPrefix())
	assert.NilError(t, err)
	for _, kv := range res.Kvs {
		t.Log(string(kv.Key), string(kv.Value))
	}

	_, err = client.Delete(context.TODO(), node1.GetValueKey())
	assert.NilError(t, err)
	res, err = client.Get(context.TODO(), funcName, clientv3.WithPrefix())
	assert.NilError(t, err)
	for _, kv := range res.Kvs {
		t.Log(string(kv.Key), string(kv.Value))
	}

	_, err = client.Put(context.TODO(), node1.GetValueKey(), string(node1.Bytes()))
	assert.NilError(t, err)
	_, err = client.Delete(context.TODO(), funcName, clientv3.WithPrefix())
	assert.NilError(t, err)
	res, err = client.Get(context.TODO(), funcName, clientv3.WithPrefix())
	assert.NilError(t, err)
	for _, kv := range res.Kvs {
		t.Log(string(kv.Key), string(kv.Value))
	}
}
