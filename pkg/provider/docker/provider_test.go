package docker

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"
	_ "github.com/miRemid/cqless/pkg/utils"
	"gotest.tools/v3/assert"
)

var (
	p = newProvider()
)

func init() {
	if err := p.Init(types.GetConfig().Gateway); err != nil {
		panic(err)
	}
	if err := cninetwork.Init(types.GetConfig()); err != nil {
		panic(err)
	}
}

func Test_Inspect(t *testing.T) {

	id := "90359a8e77a9bf829628274c96e81f24be31548f08ea74ba27f0120e8d221360"

	data, err := p.Inspect(context.Background(), id)
	assert.NilError(t, err)
	t.Log(data.ID)
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
		Image: "nginx:alpine",
		Name:  "nginx",
	}, cninetwork.DefaultManager)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(fn.IPAddress)
}

func Test_RemoveImage(t *testing.T) {
	_, err := p.Remove(context.Background(), types.FunctionRemoveRequest{
		FunctionName: "nginx",
	}, cninetwork.DefaultManager)
	assert.NilError(t, err)
}

func Test_Resolve(t *testing.T) {

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
		_, err = p.Remove(ctx, types.FunctionRemoveRequest{
			FunctionName: "nginx",
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
