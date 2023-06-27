package docker

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/rs/zerolog/log"
	"gotest.tools/v3/assert"
)

var (
	p = newProvider()
)

func init() {
	if err := p.Init(types.GetConfig().Provider, log.Logger); err != nil {
		panic(err)
	}
	if err := cninetwork.Init(types.GetConfig().Network); err != nil {
		panic(err)
	}
}

func Test_PullImage(t *testing.T) {
	t.Log("Connect to docker, prepare to pull the nginx:alpine")
	err := p.pull(context.Background(), types.FunctionCreateRequest{
		Image: "nginx:alpine",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func Test_DeployImage(t *testing.T) {
	t.Log("Connect to docker, prepare to pull the nginx:alpine\n")
	fn, err := p.Deploy(context.TODO(), types.FunctionCreateRequest{
		Image: "kamir3mid/test",
		Name:  "test",
	}, cninetwork.DefaultManager)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(fn.IPAddress)
}

func Test_RemoveImage(t *testing.T) {
	err := p.Remove(context.Background(), types.FunctionRemoveRequest{
		FunctionRequest: types.FunctionRequest{
			FunctionName: "nginx",
		},
	}, cninetwork.DefaultManager)
	assert.NilError(t, err)
}

func Test_Inspect(t *testing.T) {
	fn, err := p.Inspect(context.Background(), types.FunctionInspectRequest{
		FunctionRequest: types.FunctionRequest{
			FunctionName: "nginx",
		},
	}, cninetwork.DefaultManager)
	assert.NilError(t, err)
	t.Log(fn)
	fns, err := p.Inspect(context.Background(), types.FunctionInspectRequest{}, cninetwork.DefaultManager)
	assert.NilError(t, err)
	t.Log(fns)
}

func Test_Deploy_Inspect_Remove(t *testing.T) {
	t.Log("Connect to docker, prepare to pull the nginx:alpine\n")
	ctx := context.Background()
	fn, err := p.Deploy(ctx, types.FunctionCreateRequest{
		Image: "nginx:alpine",
		Name:  "nginx",
	}, cninetwork.DefaultManager)
	assert.NilError(t, err)
	defer func() {
		err = p.Remove(ctx, types.FunctionRemoveRequest{
			FunctionRequest: types.FunctionRequest{
				FunctionName: "nginx",
			},
		}, cninetwork.DefaultManager)
		assert.NilError(t, err)
	}()
	u, err := p.Resolve(ctx, "nginx", cninetwork.DefaultManager)
	assert.NilError(t, err)
	t.Log(u)
	command := exec.Command("curl", "http://"+fn.IPAddress)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	assert.NilError(t, command.Run())
}
