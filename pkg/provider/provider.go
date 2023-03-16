package provider

import (
	gocni "github.com/containerd/go-cni"
)

type ProviderInterface interface {
	Pull(imgName string) error
	Deploy(gocni.CNI) error
	Pause() error
	Remove() error
	Status() error
}
