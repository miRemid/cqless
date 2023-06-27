package resolver

import (
	"context"
	"testing"

	"github.com/miRemid/cqless/pkg/resolver/types"
	"gotest.tools/v3/assert"
)

var (
	opt = &types.ResolverOption{
		StorageOption: &types.StorageOption{
			Strategy: types.STORAGE_LOCAL,
		},
		SelectorOption: &types.SelectorOption{
			Strategy: types.SELECTOR_RANDOM,
		},
	}
	funcName = "test"
	node1    = types.NewNode("http", "127.0.0.1:8080", funcName, nil)
	node2    = types.NewNode("wss", "127.0.0.1:8081", funcName, nil)
)

func Test_init(t *testing.T) {
	assert.NilError(t, Init(opt))
}

func Test_Add(t *testing.T) {
	assert.NilError(t, Init(opt))
	assert.NilError(t, Register(context.Background(), funcName, node1))
	assert.NilError(t, Register(context.Background(), funcName, node1))
	assert.NilError(t, Register(context.Background(), funcName, node2))

	for i := 0; i < 10; i++ {
		node, err := Next(context.Background(), funcName)
		assert.NilError(t, err)
		t.Log(node)
	}

}
