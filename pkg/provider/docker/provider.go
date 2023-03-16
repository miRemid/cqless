package docker

import (
	"context"

	gocni "github.com/containerd/go-cni"
)

type Provider struct {
}

func (p *Provider) Deploy(cni gocni.CNI) error {
	cni.Setup(
		context.Background(), "11", "111",
	)
	return nil
}
