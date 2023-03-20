package docker

import (
	"os"
	"path"
	"strings"

	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"github.com/miRemid/cqless/pkg/types"
)

func (p *DockerProvider) convertEnvStringsToMap(envs []string) map[string]string {

	var env = make(map[string]string)
	for _, e := range envs {
		splits := strings.Split(e, "=")
		env[splits[0]] = splits[1]
	}
	return env

}

func (p *DockerProvider) createFunction(info dtypes.ContainerJSON, fnName string) *types.Function {
	var fn = new(types.Function)
	fn.ID = info.ID
	fn.PID = uint32(info.State.Pid)
	fn.Name = fnName
	fn.EnvVars = p.convertEnvStringsToMap(info.Config.Env)
	fn.Metadata = info.Config.Labels
	return fn
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
