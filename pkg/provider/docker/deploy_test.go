package docker

import (
	"context"
	"testing"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/config"
	"github.com/miRemid/cqless/pkg/types"
)

var test_config = config.NetworkConfig{
	BinaryPath:      "/opt/cni/bin",
	ConfigPath:      "/opt/cni/net.d",
	ConfigFileName:  "10-cqless.conflist",
	NetworkSavePath: "/opt/cni/net.d",

	IfPrefix:    "cqeth",
	NetworkName: "cqless-cni-bridge",
	BridgeName:  "cqless0",
	SubNet:      "10.72.0.0/16",
}

func Test_PullImage(t *testing.T) {
	p := NewProvider()
	p.Init()
	t.Log("Connect to docker, prepare to pull the nginx:alpine")
	err := p.pull(context.Background(), types.FunctionDeployRequest{
		Image: "nginx:alpine",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_DeployImage(t *testing.T) {

	p := NewProvider()
	manager := new(cninetwork.CNIManager)
	manager.InitNetwork(&test_config)

	p.Init()
	t.Log("Connect to docker, prepare to pull the nginx:alpine\n")
	err := p.Deploy(context.TODO(), types.FunctionDeployRequest{
		Image:   "nginx:alpine",
		Service: "nginx",
	}, manager, types.DEFAULT_FUNCTION_NAMESPACE, true)
	if err != nil {
		t.Fatal(err)
	}
}
