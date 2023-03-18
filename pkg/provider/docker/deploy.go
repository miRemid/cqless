package docker

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"

	dtypes "github.com/docker/docker/api/types" // docker types
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"k8s.io/apimachinery/pkg/api/resource"
)

// type pullResponse struct {
// 	Status         string `json:"status"`
// 	ID             string `json:"id,omitempty"`
// 	Error          string `json:"error"`
// 	Progress       string `json:"progress"`
// 	ProgressDetail struct {
// 		Current int `json:"current"`
// 		Total   int `json:"total"`
// 	} `json:"progressDetail"`
// }

// func (r *pullResponse) String() string {
// 	// status:
// 	format := r.Status
// 	if r.ID != "" {
// 		format += " " + "id: " + r.ID
// 	} else {
// 		format += "\r"
// 		return format
// 	}
// 	if r.ProgressDetail != nil {
// 		format += fmt.Sprintf(" %d/%d", r.ProgressDetail.Current, r.ProgressDetail.Total)
// 	}
// 	return format + "\r"
// }

func (p *DockerProvider) pull(ctx context.Context, req types.FunctionDeployRequest, alwaysPull bool) error {
	body, err := p.cli.ImagePull(ctx, req.Image, dtypes.ImagePullOptions{})
	if err != nil {
		return err
	}
	// buffer := bytes.NewBuffer(data)
	defer body.Close()
	// TODO: 进度条
	// reader := bufio.NewReader(body)
	// for {
	// 	data, err := reader.ReadBytes('\n')
	// 	if err != nil {
	// 		if err != io.EOF {
	// 			return err
	// 		}
	// 		break
	// 	}
	// 	pullResponse := new(pullResponse)
	// 	if err := json.Unmarshal(data, pullResponse); err != nil {
	// 		return err
	// 	}
	// 	fmt.Println(pullResponse)
	// }
	// io.Copy(os.Stdout, body)
	return nil
}

func (p *DockerProvider) Deploy(ctx context.Context, req types.FunctionDeployRequest, cni *cninetwork.CNIManager, namespace string, alwaysPull bool) error {
	fmt.Printf("start to pull %s\n", req.Image)
	err := p.pull(ctx, req, alwaysPull)
	if err != nil {
		return err
	}
	labels, err := req.BuildLabels()
	if err != nil {
		return err
	}
	envs := req.BuildEnv()
	mounts := p.getOSMounts()

	var containerResources container.Resources

	if req.Limits != nil && len(req.Limits.Memory) > 0 {
		qty, err := resource.ParseQuantity(req.Limits.Memory)
		if err != nil {
			log.Printf("error parsing (%q) as quantity: %s", req.Limits.Memory, err.Error())
		}
		containerResources.Memory = qty.Value()
	}

	resp, err := p.cli.ContainerCreate(ctx,
		&container.Config{
			Env:      envs,
			Labels:   labels,
			Hostname: req.Service,
			Image:    req.Image,
		},
		&container.HostConfig{
			Mounts:      mounts,
			Resources:   containerResources,
			NetworkMode: "none", // 我们将使用cni来为container提供网络
		},
		nil, nil, req.Service)
	if err != nil {
		return err
	}
	if err := p.cli.ContainerStart(ctx, resp.ID, dtypes.ContainerStartOptions{}); err != nil {
		return err
	}
	info, err := p.cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return err
	}
	return p.createNetwork(ctx, info, cni)
}

// 为容器创建一个CNI网络用于通信
func (p *DockerProvider) createNetwork(ctx context.Context, container dtypes.ContainerJSON, cni *cninetwork.CNIManager) error {
	labels := map[string]string{}
	_, err := cni.CreateCNINetwork(ctx, types.Container{
		ID:           container.ID,
		PID:          uint32(container.State.Pid),
		Name:         container.Name,
		NetNamespace: container.NetworkSettings.SandboxKey,
	}, labels)
	if err != nil {
		return err
	}
	return nil
}

// Mount 宿主机的DNS信息和Hosts信息到容器中
func (p *DockerProvider) getOSMounts() []mount.Mount {
	hostsDir := "/var/lib/cqless"
	if v, ok := os.LookupEnv("hosts_dir"); ok && len(v) > 0 {
		hostsDir = v
	}

	mounts := []mount.Mount{}
	mounts = append(mounts, mount.Mount{
		Target: "/etc/resolv.conf",
		Type:   "bind",
		Source: path.Join(hostsDir, "resolv.conf"),
		BindOptions: &mount.BindOptions{
			Propagation: mount.PropagationRPrivate,
		},
		ReadOnly: true,
	})

	mounts = append(mounts, mount.Mount{
		Target: "/etc/hosts",
		Type:   "bind",
		Source: path.Join(hostsDir, "hosts"),
		BindOptions: &mount.BindOptions{
			Propagation: mount.PropagationRPrivate,
		},
		ReadOnly: true,
	})
	return mounts
}
