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

func Test_InitAndUninstallNetwork(t *testing.T) {
	err := InitNetwork(&test_config)
	assert.NilError(t, err)
}

func Test_CreateAndDeleteCNINetwork(t *testing.T) {
	err := InitNetwork(&test_config)
	assert.NilError(t, err)
	labels := map[string]string{}
	ID := "64c22e80ac91ec956e63815c26df7b7aec73153936643bd78cfbadb0c6272f6d"
	PID := 172363
	c := types.Container{
		ID:           ID,
		PID:          uint32(PID),
		Name:         "Test",
		NetNamespace: "/var/run/netns/test",
	}
	_, err = CreateCNINetwork(context.Background(), c, labels)
	assert.NilError(t, err)
	ip, err := GetIPAddress(ID, uint32(PID))
	assert.NilError(t, err)
	t.Log(ip)
	err = DeleteCNINetwork(context.TODO(), c)
	assert.NilError(t, err)
	assert.NilError(t, Uninstall())
}
