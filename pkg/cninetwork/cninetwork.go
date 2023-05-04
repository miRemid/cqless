// Modifed from https://github.com/openfaas/faasd/blob/master/pkg/cninetwork/cni_network.go
package cninetwork

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	gocni "github.com/containerd/go-cni"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var cniconf = `
{
    "cniVersion": "0.4.0",
    "name": "%s",
    "plugins": [
      {
        "type": "bridge",
        "bridge": "%s",
        "isGateway": true,
        "ipMasq": true,
        "ipam": {
            "type": "host-local",
            "subnet": "%s",
            "dataDir": "%s",
            "routes": [
                { "dst": "0.0.0.0/0" }
            ]
        }
      },
      {
        "type": "firewall"
      }
    ]
}
`

var (
	DefaultManager *CNIManager
)

func init() {
	DefaultManager = new(CNIManager)
}

type CNIManager struct {
	cli    gocni.CNI
	config *types.NetworkConfig
}

func (m *CNIManager) GenerateJSON() []byte {
	return []byte(fmt.Sprintf(
		cniconf,
		m.config.NetworkName,
		m.config.BridgeName,
		m.config.SubNet,
		m.config.NetworkSavePath))
}

// InitNetwork initialize the default cni network for all
// function containers
func (m *CNIManager) Init(config *types.CQLessConfig) error {
	m.config = config.Network

	if !dirExists(m.config.ConfigPath) {
		if err := os.MkdirAll(m.config.ConfigPath, 0755); err != nil {
			return err
		}
	}
	netConfig := path.Join(m.config.ConfigPath, m.config.ConfigFileName)
	if err := os.WriteFile(netConfig, m.GenerateJSON(), 0644); err != nil {
		return err
	}

	cni, err := gocni.New(
		gocni.WithPluginConfDir(m.config.ConfigPath),
		gocni.WithPluginDir([]string{m.config.BinaryPath}),
		gocni.WithInterfacePrefix(m.config.IfPrefix),
	)
	if err != nil {
		return err
	}
	if err := cni.Load(gocni.WithLoNetwork, gocni.WithConfListFile(filepath.Join(m.config.ConfigPath, m.config.ConfigFileName))); err != nil {
		return err
	}
	m.cli = cni
	return nil
}

// InitNetwork initialize the default cni manager
func Init(config *types.CQLessConfig) error {
	return DefaultManager.Init(config)
}

// DeleteCNINetwork deletes a CNI network based on container's id and pid
func (m *CNIManager) DeleteCNINetwork(ctx context.Context, fn *types.Function) error {
	log.Printf("[Delete] removing CNI network for: %s\n", fn.ID)

	id := m.NetID(fn.ID, fn.PID)
	netns := m.NetNamespace(fn)

	if err := m.cli.Remove(ctx, id, netns); err != nil {
		return errors.Wrapf(err, "Failed to remove network for task: %q, %v", id, err)
	}
	log.Printf("[Delete] removed: %s from namespace: %s, ID: %s\n", fn.Name, netns, id)

	return nil
}

func DeleteCNINetwork(ctx context.Context, fn *types.Function) error {
	return DefaultManager.DeleteCNINetwork(ctx, fn)
}

// CreateCNINetwork creates a CNI network interface and attaches it to the context
func (m *CNIManager) CreateCNINetwork(ctx context.Context, fn *types.Function) (*gocni.Result, error) {
	log.Info().Msgf("start to create cninetwork for function: %s", fn.Name)
	id := m.NetID(fn.ID, fn.PID)
	netns := m.NetNamespace(fn)
	result, err := m.cli.Setup(ctx, id, netns, gocni.WithLabels(fn.Labels))
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, errors.Wrapf(err, "Failed to setup network for task %q: %v", id, err)
	}
	ipAddress, err := m.GetIPAddress(fn)
	if err != nil {
		return nil, err
	}
	fn.IPAddress = ipAddress
	return result, nil
}

func CreateCNINetwork(ctx context.Context, fn *types.Function) (*gocni.Result, error) {
	return DefaultManager.CreateCNINetwork(ctx, fn)
}

// GetIPAddress returns the IP address from container based on container name and PID
func (m *CNIManager) GetIPAddress(fn *types.Function) (string, error) {
	return m.GetIPAddressRaw(fn.ID, fn.PID)
}
func (m *CNIManager) GetIPAddressRaw(container string, PID uint32) (string, error) {
	CNIDir := path.Join(m.config.NetworkSavePath, m.config.NetworkName)
	log.Debug().Msgf("search CNIDir: %s", CNIDir)
	files, err := os.ReadDir(CNIDir)
	if err != nil {
		return "", fmt.Errorf("failed to read CNI dir for container %s: %v", container, err)
	}

	for _, file := range files {
		// each fileName is an IP address
		fileName := file.Name()

		resultsFile := filepath.Join(CNIDir, fileName)
		found, err := m.isCNIResultForPID(resultsFile, container, PID)
		if err != nil {
			return "", err
		}

		if found {
			return fileName, nil
		}
	}

	return "", fmt.Errorf("unable to get IP address for container: %s", container)
}
func GetIPAddress(fn *types.Function) (string, error) {
	return DefaultManager.GetIPAddress(fn)
}
func GetIPAddressRaw(container string, PID uint32) (string, error) {
	return DefaultManager.GetIPAddressRaw(container, PID)
}

// CNIGateway returns the gateway for default subnet
func (m *CNIManager) CNIGateway() (string, error) {
	ip, _, err := net.ParseCIDR(m.config.SubNet)
	if err != nil {
		return "", fmt.Errorf("error formatting gateway for network %s", m.config.SubNet)
	}
	ip = ip.To4()
	ip[3] = 1
	return ip.String(), nil
}

func CNIGateway() (string, error) {
	return DefaultManager.CNIGateway()
}

// isCNIResultForPID confirms if the CNI result file contains the
// process name, PID and interface name
//
// Example:
//
// /var/run/cni/openfaas-cni-bridge/10.62.0.2
//
// nats-621
// eth1
func (m *CNIManager) isCNIResultForPID(fileName, container string, PID uint32) (bool, error) {
	found := false
	f, err := os.Open(fileName)
	if err != nil {
		return false, fmt.Errorf("failed to open CNI IP file for %s: %v", fileName, err)
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	processLine, _ := reader.ReadString('\n')
	if strings.Contains(processLine, fmt.Sprintf("%s-%d", container, PID)) {
		ethNameLine, _ := reader.ReadString('\n')
		if strings.Contains(ethNameLine, m.config.IfPrefix) {
			found = true
		}
	}
	return found, nil
}

// NetID generates the network IF based on task name and task PID
func (m *CNIManager) NetID(id string, pid uint32) string {
	return fmt.Sprintf("%s-%d", id, pid)
}

// NetNamespace generates the namespace path based on task PID.
func (m *CNIManager) NetNamespace(fn *types.Function) string {
	if len(fn.Namespace) > 0 {
		return fn.Namespace
	}
	fn.Namespace = fmt.Sprintf(m.config.NamespaceFormat, fn.ID)
	return fn.Namespace
}

// Uninstall 删除所有网络
// TODO: 完成后续工作
func (m *CNIManager) Uninstall() error {
	// 1. 删除所有网络
	// 2. 删除网桥
	// 由于在go-cni的文档中没有明确找到有关删除网桥的接口，因此使用命令行的方式删除
	// 需要安装net-tools和brctl工具包
	// 2.1 关闭网桥 ipconfig $BRIDGE_NAME down
	command := exec.Command("ifconfig", m.config.BridgeName, "down")
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return err
	}
	// 2.2 删除网桥 brctl delbr $BRIDGE_NAME
	command = exec.Command("brctl", "delbr", m.config.BridgeName)
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return err
	}
	return nil
}

func Uninstall() error {
	return DefaultManager.Uninstall()
}

func dirExists(dirname string) bool {
	exists, info := pathExists(dirname)
	if !exists {
		return false
	}

	return info.IsDir()
}

func pathExists(path string) (bool, os.FileInfo) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}

	return true, info
}
