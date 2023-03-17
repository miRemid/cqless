// Modifed from https://github.com/openfaas/faasd/blob/master/pkg/cninetwork/cni_network.go
package cninetwork

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"

	gocni "github.com/containerd/go-cni"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/pkg/errors"
)

var (
	defaultManager *CNIManager
)

func init() {
	defaultManager = new(CNIManager)
}

type CNIManager struct {
	cli    gocni.CNI
	config CNIConfig
}

// InitNetwork initialize the default cni network for all
// function containers
func (m *CNIManager) InitNetwork(config CNIConfig) error {
	m.config = config

	if !dirExists(config.ConfDir) {
		if err := os.MkdirAll(config.ConfDir, 0755); err != nil {
			return err
		}
	}
	netConfig := path.Join(config.ConfDir, config.ConfFileName)
	if err := os.WriteFile(netConfig, config.GenerateJSON(), 0644); err != nil {
		return err
	}

	cni, err := gocni.New(
		gocni.WithPluginConfDir(config.ConfDir),
		gocni.WithPluginDir([]string{config.BinDir}),
		gocni.WithInterfacePrefix(config.IfPrefix),
	)
	if err != nil {
		return err
	}
	if err := cni.Load(gocni.WithLoNetwork, gocni.WithConfListFile(filepath.Join(config.ConfDir, config.ConfFileName))); err != nil {
		return err
	}
	m.cli = cni
	return nil
}

// InitNetwork initialize the default cni manager
func InitNetwork(config CNIConfig) error {
	return defaultManager.InitNetwork(config)
}

// DeleteCNINetwork deletes a CNI network based on container's id and pid
func (m *CNIManager) DeleteCNINetwork(ctx context.Context, cnic types.Container) error {
	log.Printf("[Delete] removing CNI network for: %s\n", cnic.ID)

	id := m.NetID(cnic.ID, cnic.PID)
	netns := m.NetNamespace(cnic)

	if err := m.cli.Remove(ctx, id, netns); err != nil {
		return errors.Wrapf(err, "Failed to remove network for task: %q, %v", id, err)
	}
	log.Printf("[Delete] removed: %s from namespace: %s, ID: %s\n", cnic.Name, netns, id)

	return nil
}

func DeleteCNINetwork(ctx context.Context, cninc types.Container) error {
	return defaultManager.DeleteCNINetwork(ctx, cninc)
}

// CreateCNINetwork creates a CNI network interface and attaches it to the context
func (m *CNIManager) CreateCNINetwork(ctx context.Context, cnic types.Container, labels map[string]string) (*gocni.Result, error) {
	id := m.NetID(cnic.ID, cnic.PID)
	netns := m.NetNamespace(cnic)
	result, err := m.cli.Setup(ctx, id, netns, gocni.WithLabels(labels))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to setup network for task %q: %v", id, err)
	}
	return result, nil
}

func CreateCNINetwork(ctx context.Context, cnic types.Container, labels map[string]string) (*gocni.Result, error) {
	return defaultManager.CreateCNINetwork(ctx, cnic, labels)
}

// GetIPAddress returns the IP address from container based on container name and PID
func (m *CNIManager) GetIPAddress(container string, PID uint32) (string, error) {
	CNIDir := path.Join(m.config.DataDir, m.config.NetworkName)

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

func GetIPAddress(container string, PID uint32) (string, error) {
	return defaultManager.GetIPAddress(container, PID)
}

// CNIGateway returns the gateway for default subnet
func (m *CNIManager) CNIGateway() (string, error) {
	ip, _, err := net.ParseCIDR(m.config.Subnet)
	if err != nil {
		return "", fmt.Errorf("error formatting gateway for network %s", m.config.Subnet)
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
func (m *CNIManager) NetNamespace(cnic types.Container) string {
	if len(cnic.NetNamespace) > 0 {
		return cnic.NetNamespace
	}
	return fmt.Sprintf(m.config.NamespaceFmt, cnic.ID)
}

func dirEmpty(dirname string) (isEmpty bool) {
	if !dirExists(dirname) {
		return
	}

	f, err := os.Open(dirname)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	// If the first file is EOF, the directory is empty
	if _, err = f.Readdir(1); err == io.EOF {
		isEmpty = true
	}
	return isEmpty
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
