package cninetwork

import (
	"context"
	"os/exec"
	"testing"

	"github.com/miRemid/cqless/v1/pkg/pb"
	"github.com/miRemid/cqless/v1/pkg/types"
	"gotest.tools/v3/assert"
)

func Test_GetBridge(t *testing.T) {
	// 获取网桥
	command := exec.Command("ifconfig", "docker0")
	if err := command.Run(); err != nil {
		t.Fatal(err)
	}
}

func Test_GetAddress(t *testing.T) {
	if err := Init(types.GetConfig().Network); err != nil {
		panic(err)
	}
	c := &pb.Function{
		Id:        "a0cb7092362a4ebb7f68dfb63fe41da70c8b07fe1dc490eb154241c5cee9d8e1",
		Pid:       204516,
		Name:      "nginx",
		Namespace: "/var/run/docker/netns/05067a2894b3",
	}

	ip, err := GetIPAddress(c)
	assert.NilError(t, err)
	t.Log(ip)
}

func Test_CreateAndDeleteCNINetwork(t *testing.T) {
	if err := Init(types.GetConfig().Network); err != nil {
		panic(err)
	}
	ID := "9a81254df505249b9c9489aad08f94a39f0c0a768c3f19b5365444725ed52452"
	PID := 77796
	c := &pb.Function{
		Id:        ID,
		Pid:       int64(PID),
		Name:      "Nginx",
		Namespace: "/var/run/docker/netns/6fe11047e24f",
	}
	_, err := CreateCNINetwork(context.Background(), c)
	assert.NilError(t, err)
	ip, err := GetIPAddress(c)
	assert.NilError(t, err)
	t.Log(ip)
	err = DeleteCNINetwork(context.TODO(), c)
	assert.NilError(t, err)
	assert.NilError(t, Uninstall())
}
