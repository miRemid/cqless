package docker

import (
	"context"
	"fmt"

	"github.com/miRemid/cqless/pkg/types"

	dtypes "github.com/docker/docker/api/types" // docker types
	"github.com/docker/docker/api/types/container"
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

func (p *DockerProvider) Deploy(ctx context.Context, req types.FunctionCreateRequest) (*types.Function, error) {
	fmt.Printf("start to pull %s\n", req.Image)
	err := p.pull(ctx, req)
	if err != nil {
		return nil, err
	}
	labels, err := req.BuildLabels()
	if err != nil {
		return nil, err
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
			Hostname: req.Name,
			Image:    req.Image,
		},
		&container.HostConfig{
			Mounts:      mounts,
			Resources:   containerResources,
			NetworkMode: "none", // 我们将使用cni来为container提供网络
		},
		nil, nil, req.Name)
	if err != nil {
		return nil, err
	}
	if err := p.cli.ContainerStart(ctx, resp.ID, dtypes.ContainerStartOptions{}); err != nil {
		return nil, err
	}
	info, err := p.cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, err
	}
	fn := p.createFunction(info, req.Name)
	return fn, nil
}
