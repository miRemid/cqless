// Modifed from https://github.com/openfaas/faasd/blob/master/pkg/cninetwork/cni_network.go
package cninetwork

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	gocni "github.com/containerd/go-cni"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/pkg/errors"
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
	defaultManager *CNIManager
)

func init() {
	defaultManager = new(CNIManager)
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
func (m *CNIManager) InitNetwork(config *types.NetworkConfig) error {
	m.config = config

	if !dirExists(config.ConfigPath) {
		if err := os.MkdirAll(config.ConfigPath, 0755); err != nil {
			return err
		}
	}
	netConfig := path.Join(config.ConfigPath, config.ConfigFileName)
	if err := os.WriteFile(netConfig, m.GenerateJSON(), 0644); err != nil {
		return err
	}

	cni, err := gocni.New(
		gocni.WithPluginConfDir(config.ConfigPath),
		gocni.WithPluginDir([]string{config.BinaryPath}),
		gocni.WithInterfacePrefix(config.IfPrefix),
	)
	if err != nil {
		return err
	}
	if err := cni.Load(gocni.WithLoNetwork, gocni.WithConfListFile(filepath.Join(config.ConfigPath, config.ConfigFileName))); err != nil {
		return err
	}
	m.cli = cni
	return nil
}

// InitNetwork initialize the default cni manager
func InitNetwork(config *types.NetworkConfig) error {
	return defaultManager.InitNetwork(config)
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
	return defaultManager.DeleteCNINetwork(ctx, fn)
}

// CreateCNINetwork creates a CNI network interface and attaches it to the context
func (m *CNIManager) CreateCNINetwork(ctx context.Context, fn *types.Function) (*gocni.Result, error) {
	id := m.NetID(fn.ID, fn.PID)
	netns := m.NetNamespace(fn)
	result, err := m.cli.Setup(ctx, id, netns, gocni.WithLabels(fn.Metadata))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to setup network for task %q: %v", id, err)
	}
	ipAddress, _ := m.GetIPAddress(fn)
	fn.IPAddress = ipAddress
	return result, nil
}

func CreateCNINetwork(ctx context.Context, fn *types.Function) (*gocni.Result, error) {
	return defaultManager.CreateCNINetwork(ctx, fn)
}

// GetIPAddress returns the IP address from container based on container name and PID
func (m *CNIManager) GetIPAddress(fn *types.Function) (string, error) {
	return m.GetIPAddressRaw(fn.Name, fn.PID)
}
func (m *CNIManager) GetIPAddressRaw(container string, PID uint32) (string, error) {
	CNIDir := path.Join(m.config.NetworkSavePath, m.config.NetworkName)

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
	return defaultManager.GetIPAddress(fn)
}
func GetIPAddressRaw(container string, PID uint32) (string, error) {
	return defaultManager.GetIPAddressRaw(container, PID)
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
	return defaultManager.CNIGateway()
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
	fmt.Println(command)
	if err := command.Run(); err != nil {
		return err
	}
	// 2.2 删除网桥 brctl delbr $BRIDGE_NAME
	command = exec.Command("brctl", "delbr", m.config.BridgeName)
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	fmt.Println(command)
	if err := command.Run(); err != nil {
		return err
	}
	return nil
}

func Uninstall() error {
	return defaultManager.Uninstall()
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
