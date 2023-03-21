package docker

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"
	"gotest.tools/v3/assert"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

var test_config = &types.NetworkConfig{
	BinaryPath:      "/opt/cni/bin",
	ConfigPath:      "/opt/cni/net.d",
	ConfigFileName:  "10-cqless.conflist",
	NetworkSavePath: "/opt/cni/net.d",

	IfPrefix:    "cqeth",
	NetworkName: "cqless-cni-bridge",
	BridgeName:  "cqless0",
	SubNet:      "10.72.0.0/16",
}

func Test_Inspect(t *testing.T) {

	id := "1cc6a820e3695d9da7d85861a1e4e7e9b269678bdb60f42179e9f867ffb62b00"

	data, err := p.Inspect(context.Background(), id)
	assert.NilError(t, err)
	log.Info(data.Config.Env)

}

func Test_PullImage(t *testing.T) {
	p := NewProvider()
	p.Init()
	t.Log("Connect to docker, prepare to pull the nginx:alpine")
	err := p.pull(context.Background(), types.FunctionCreateRequest{
		Image: "nginx:alpine",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func Test_DeployImage(t *testing.T) {

	p := NewProvider()
	manager := new(cninetwork.CNIManager)
	manager.InitNetwork(test_config)

	p.Init()
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
	p := NewProvider()
	p.Init()
	manager := new(cninetwork.CNIManager)
	manager.InitNetwork(test_config)

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
