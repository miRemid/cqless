package docker

import (
	"context"
	"os"
	"path"
	"strings"
	"sync"

	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/rs/zerolog/log"
)

func (p *DockerProvider) convertEnvStringsToMap(envs []string) map[string]string {

	var env = make(map[string]string)
	for _, e := range envs {
		splits := strings.Split(e, "=")
		env[splits[0]] = splits[1]
	}
	return env

}

func (p *DockerProvider) createFunction(info dtypes.ContainerJSON) *types.Function {
	var fn = new(types.Function)

	fn.ID = info.ID
	fn.PID = uint32(info.State.Pid)
	fn.Name = info.Config.Labels[types.DEFAULT_FUNCTION_NAME_LABEL]
	fn.FullName = info.Name
	fn.EnvVars = p.convertEnvStringsToMap(info.Config.Env)
	fn.Metadata = info.Config.Labels
	fn.Namespace = info.NetworkSettings.SandboxKey
	fn.Status = info.State.Status
	return fn
}

func (p *DockerProvider) getAllFunctionContainers(ctx context.Context, fs ...filters.KeyValuePair) ([]dtypes.Container, error) {
	filter := filters.NewArgs(fs...)
	filter.Add("label", types.DEFAULT_FUNCTION_NAME_LABEL)
	containers, err := p.cli.ContainerList(ctx, dtypes.ContainerListOptions{
		Filters: filter,
	})
	return containers, err
}

func (p *DockerProvider) getAllFunctions(ctx context.Context, cni *cninetwork.CNIManager, fs ...filters.KeyValuePair) ([]*types.Function, error) {
	containers, err := p.getAllFunctionContainers(ctx, fs...)
	if err != nil {
		return nil, err
	}
	var functionChan = make(chan *types.Function, len(containers))
	wg := sync.WaitGroup{}
	for _, info := range containers {
		wg.Add(1)
		go func(c dtypes.Container) {
			defer wg.Done()
			function, err := p.getFunctionByContainer(ctx, c, cni)
			if err != nil {
				return
			}
			functionChan <- function
		}(info)
	}
	wg.Wait()
	close(functionChan)
	var res = make([]*types.Function, 0)
	for len(functionChan) != 0 {
		fn := <-functionChan
		res = append(res, fn)
	}
	return res, nil
}

func (p *DockerProvider) getAllFunctionsByName(ctx context.Context, fnName string, cni *cninetwork.CNIManager) ([]*types.Function, error) {
	functions, err := p.getAllFunctions(ctx, cni, filters.Arg("name", fnName))
	if err != nil {
		return nil, err
	}
	return functions, nil
}

func (p *DockerProvider) getFunctionByContainer(ctx context.Context, c dtypes.Container, cni *cninetwork.CNIManager) (*types.Function, error) {
	log.Debug().Str("getFunctionByContainer.containerID", c.ID).Send()
	info, err := p.cli.ContainerInspect(ctx, c.ID)
	if err != nil {
		log.Err(err).Send()
		return nil, err
	}
	function := p.createFunction(info)
	ip, err := cni.GetIPAddress(function)
	if err != nil {
		return nil, err
	}
	function.IPAddress = ip
	return function, nil
}

// Mount 宿主机的DNS信息和Hosts信息到容器中
func (p *DockerProvider) getOSMounts() []mount.Mount {
	hostsDir := types.DEFAULT_CONFIG_PATH
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
