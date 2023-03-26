package cninetwork

import (
	"context"
	"os/exec"
	"testing"

	"github.com/miRemid/cqless/pkg/types"
	"github.com/sirupsen/logrus"
	"gotest.tools/v3/assert"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

var test_config = types.Network{
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
	c := &types.Function{
		ID:        "a0cb7092362a4ebb7f68dfb63fe41da70c8b07fe1dc490eb154241c5cee9d8e1",
		PID:       204516,
		Name:      "nginx",
		Namespace: "/var/run/docker/netns/05067a2894b3",
	}

	ip, err := GetIPAddress(c)
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
	ID := "253713407df3bce594d37cbd4d7db5a08d0cc7affe16f673d1b535a04c0bd98e"
	PID := 556512
	c := &types.Function{
		ID:        ID,
		PID:       uint32(PID),
		Name:      "Nginx",
		Namespace: "/var/run/docker/netns/44d1f3f69675",
	}
	_, err = CreateCNINetwork(context.Background(), c)
	assert.NilError(t, err)
	ip, err := GetIPAddress(c)
	assert.NilError(t, err)
	t.Log(ip)
	err = DeleteCNINetwork(context.TODO(), c)
	assert.NilError(t, err)
	assert.NilError(t, Uninstall())
}
