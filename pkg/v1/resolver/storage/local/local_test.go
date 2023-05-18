package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/miRemid/cqless/pkg/v1/logger"
	"github.com/miRemid/cqless/pkg/v1/resolver/types"
	"github.com/rs/zerolog/log"
	"gotest.tools/v3/assert"
)

var (
	opt = &types.StorageOption{
		Strategy: "random",
		DBPath:   "tempdata",
	}
	l        *localStorage
	funcName = "test"
	node1    = types.NewNode("http", "127.0.0.1:8080", funcName, nil)
	node2    = types.NewNode("wss", "127.0.0.1:8081", funcName, nil)
)

func init() {
	l = New(opt, log.Hook(logger.ModuleHook("storage")))
}

func Test_Register(t *testing.T) {
	assert.NilError(t, l.Init())
	assert.NilError(t, l.Register(context.Background(), "test", node1))
}

func Test_Get(t *testing.T) {
	assert.NilError(t, l.Init())
	nodes, err := l.Get(context.Background(), funcName)
	assert.NilError(t, err)
	fmt.Println(nodes)
}

func Test_UnRegister(t *testing.T) {
	assert.NilError(t, l.Init())
	assert.NilError(t, l.UnRegister(context.Background(), "test", node2))
}

func Test_Pipeline(t *testing.T) {
	assert.NilError(t, l.Init())

	assert.NilError(t, l.Register(context.Background(), "test", node1))
	assert.NilError(t, l.Register(context.Background(), "test", node2))
	assert.NilError(t, l.Register(context.Background(), "test", node1))
	nodes, err := l.Get(context.Background(), funcName)
	assert.NilError(t, err)
	fmt.Println(nodes)

	assert.NilError(t, l.UnRegister(context.Background(), "test", node1))
	nodes, err = l.Get(context.Background(), funcName)
	assert.NilError(t, err)
	fmt.Println(nodes)
}
