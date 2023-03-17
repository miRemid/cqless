package cninetwork

import (
	"context"
	"testing"

	"github.com/miRemid/cqless/pkg/types"
)

var test_config = CNIConfig{
	BinDir:       "/opt/cni/bin",
	ConfDir:      "/opt/cni/net.d",
	DataDir:      "/opt/cni/net.d",
	ConfFileName: "10-cqless.conflist",
	IfPrefix:     "cqeth",
	NetworkName:  "cqless-cni-bridge",
	BridgeName:   "cqless0",
	Subnet:       "10.72.0.0/16",
}

func Test_InitNetwork(t *testing.T) {
	err := InitNetwork(test_config)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_CreateCNINetwork(t *testing.T) {
	err := InitNetwork(test_config)
	if err != nil {
		t.Fatal(err)
	}
	labels := make(map[string]string)
	ID := "64c22e80ac91ec956e63815c26df7b7aec73153936643bd78cfbadb0c6272f6d"
	PID := 172363
	_, err = CreateCNINetwork(context.Background(), types.Container{
		ID:           ID,
		PID:          uint32(PID),
		Name:         "Test",
		NetNamespace: "/var/run/docker/netns/38b8aa22f7b3",
	}, labels)
	if err != nil {
		t.Fatal(err)
	}
	ip, err := GetIPAddress(ID, uint32(PID))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
}
