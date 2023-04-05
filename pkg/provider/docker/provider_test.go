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
}

func Test_Inspect(t *testing.T) {

	id := "90359a8e77a9bf829628274c96e81f24be31548f08ea74ba27f0120e8d221360"

	data, err := p.Inspect(context.Background(), id)
	assert.NilError(t, err)
	log.Info().Msg(data.ID)
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

	manager := new(cninetwork.CNIManager)
	assert.NilError(t, manager.InitNetwork(types.GetConfig().Network))

	t.Log("Connect to docker, prepare to pull the nginx:alpine\n")
	fn, err := p.Deploy(context.TODO(), types.FunctionCreateRequest{
		Image: "nginx:alpine",
		Name:  "nginx",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(fn.IPAddress)
}

func Test_Deploy_Inspect_Remove(t *testing.T) {
	manager := new(cninetwork.CNIManager)
	assert.NilError(t, manager.InitNetwork(types.GetConfig().Network))

	t.Log("Connect to docker, prepare to pull the nginx:alpine\n")
	ctx := context.Background()
	fn, err := p.Deploy(ctx, types.FunctionCreateRequest{
		Image: "nginx:alpine",
		Name:  "nginx",
	})
	assert.NilError(t, err)
	defer func() {
		_, err = p.Remove(ctx, types.FunctionRemoveRequest{
			FunctionName: "nginx",
		})
		assert.NilError(t, err)
	}()
	_, err = manager.CreateCNINetwork(ctx, fn)
	assert.NilError(t, err)
	defer func() {
		err = manager.DeleteCNINetwork(ctx, fn)
		assert.NilError(t, err)
	}()
	command := exec.Command("curl", "http://"+fn.IPAddress)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	assert.NilError(t, command.Run())

}
