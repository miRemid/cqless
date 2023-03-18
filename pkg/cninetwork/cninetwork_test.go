package cninetwork

import (
	"context"
	"os/exec"
	"testing"

	"github.com/miRemid/cqless/pkg/config"
	"github.com/miRemid/cqless/pkg/types"
	"gotest.tools/v3/assert"
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

func Test_GetBridge(t *testing.T) {
	// 获取网桥
	command := exec.Command("ifconfig", "docker0")
	if err := command.Run(); err != nil {
		t.Fatal(err)
	}
}

func Test_GetAddress(t *testing.T) {
	err := InitNetwork(&test_config)
	assert.NilError(t, err)
	c := types.Container{
		ID:           "a0cb7092362a4ebb7f68dfb63fe41da70c8b07fe1dc490eb154241c5cee9d8e1",
		PID:          204516,
		Name:         "nginx",
		NetNamespace: "/var/run/docker/netns/05067a2894b3",
	}

	ip, err := GetIPAddress(c.ID, c.PID)
	assert.NilError(t, err)
	t.Log(ip)
}

func Test_InitAndUninstallNetwork(t *testing.T) {
	err := InitNetwork(&test_config)
	assert.NilError(t, err)
}

func Test_CreateAndDeleteCNINetwork(t *testing.T) {
	err := InitNetwork(&test_config)
	assert.NilError(t, err)
	ID := "351898b01465113416975ffb43eb730b51c12ab9c1e1c5b5421ae20245d30f4a"
	PID := 199778
	c := types.Container{
		ID:           ID,
		PID:          uint32(PID),
		Name:         "Nginx",
		NetNamespace: "/var/run/docker/netns/4b9acd80054c",
	}
	labels := map[string]string{}
	_, err = CreateCNINetwork(context.Background(), c, labels)
	assert.NilError(t, err)
	ip, err := GetIPAddress(ID, uint32(PID))
	assert.NilError(t, err)
	t.Log(ip)
	err = DeleteCNINetwork(context.TODO(), c)
	assert.NilError(t, err)
	assert.NilError(t, Uninstall())
}
