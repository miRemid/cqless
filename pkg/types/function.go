package types

import (
	"fmt"

	"github.com/miRemid/cqless/pkg/resolver/types"
)

// Function 用于抽象provider和cninetwork之间的沟通桥梁
type Function struct {
	ID       string `json:"id"`        // 容器ID
	PID      uint32 `json:"pid"`       // 宿主机上的PID
	Name     string `json:"name"`      // 函数名称
	FullName string `json:"full_name"` // 容器名称
	Status   string `json:"status"`    // 容器状态

	Namespace    string            `json:"namespace"` // 所在Namespace
	IPAddress    string            `json:"ip"`        // CNI分配的IP地址
	WatchdogPort string            `json:"port"`      // 服务所在端口
	Scheme       string            `json:"scheme"`    // 协议
	Labels       map[string]string `json:"labels"`    // CNI bridge 用的

	EnvVars  map[string]string `json:"env"`       // 容器内的环境变量
	Metadata map[string]string `json:"meta_data"` // 容器的Meta数据
}

func (f Function) String() string {
	return fmt.Sprintf("%s %d %s", f.Name, f.PID, f.IPAddress)
}

func (f Function) Node() *types.Node {
	return types.NewNode(f.Scheme, f.IPAddress+":"+f.WatchdogPort, f.Name, f.Metadata)
}
