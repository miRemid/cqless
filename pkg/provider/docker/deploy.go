package docker

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"

	dtypes "github.com/docker/docker/api/types" // docker types
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
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

func (p *DockerProvider) pull(ctx context.Context, req types.FunctionCreateRequest) error {
	// 1. check local
	filter := filters.NewArgs(filters.Arg("reference", req.Image))
	img, err := p.cli.ImageList(ctx, dtypes.ImageListOptions{
		Filters: filter,
	})
	if err != nil {
		return err
	}
	if len(img) > 0 {
		return nil
	}
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

func (p *DockerProvider) Deploy(ctx context.Context, req types.FunctionCreateRequest, cni *cninetwork.CNIManager) (*types.Function, error) {
	p.log.Printf("start to pull %s\n", req.Image)
	err := p.pull(ctx, req)
	if err != nil {
		return nil, err
	}
	p.log.Printf("start to deploy function: %s\n", req.Name)
	labels, err := req.BuildLabels()
	if err != nil {
		return nil, err
	}
	labels[types.DEFAULT_FUNCTION_NAME_LABEL] = req.Name // 添加一个 DEFAULT_FUNCTION_NAME_LABEL = FunctionName 用于后续查找
	if req.WatchDogPort == "" {
		req.WatchDogPort = types.DEFAULT_WATCHDOG_PORT
	}
	labels[types.DEFAULT_FUNCTION_PORT_LABEL] = req.WatchDogPort // 添加一个WatchPort标签，用于Resolve

	// 计算一个Hash Label，用于区分
	hashKey := fmt.Sprintf("%s%v", req.Name, time.Now())
	hashValue := sha256.Sum256([]byte(hashKey))
	containerName := fmt.Sprintf("%s-%x", req.Name, hashValue[:4])

	envs := req.BuildEnv()
	// mounts := p.getOSMounts()

	var containerResources container.Resources

	if req.Limits != nil && len(req.Limits.Memory) > 0 {
		qty, err := resource.ParseQuantity(req.Limits.Memory)
		if err != nil {
			p.log.Printf("error parsing (%q) as quantity: %s", req.Limits.Memory, err.Error())
		}
		containerResources.Memory = qty.Value()
	}

	resp, err := p.cli.ContainerCreate(ctx,
		&container.Config{
			Env:      envs,
			Labels:   labels,
			Hostname: req.Name,
			Image:    req.Image,
		},
		&container.HostConfig{
			// Mounts:      mounts,
			Resources:   containerResources,
			NetworkMode: "none", // 我们将使用cni来为container提供网络
		},
		nil, nil, containerName)
	if err != nil {
		return nil, err
	}
	p.log.Printf("create function container, id: %s\n", resp.ID)
	if err := p.cli.ContainerStart(ctx, resp.ID, dtypes.ContainerStartOptions{}); err != nil {
		return nil, err
	}
	p.log.Printf("start function container, id: %s\n", resp.ID)
	info, err := p.cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, err
	}
	fn := p.createFunction(info)
	_, err = cni.CreateCNINetwork(ctx, fn)
	if err != nil {
		return nil, err
	}
	ip, err := cni.GetIPAddress(fn)
	if err != nil {
		return nil, err
	}
	fn.IPAddress = ip
	fn.Scheme = req.Scheme
	return fn, nil
}
