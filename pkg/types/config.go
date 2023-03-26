package types

import "time"

type CQLessConfig struct {
	Network     *Network     `json:"network"`
	Logger      *Logger      `json:"logger"`
	ProxyClient *ProxyClient `json:"proxy_client"`
}

type Network struct {
	// BinaryPath CNI插件的绝对路径
	BinaryPath string
	// ConfigPath 存放CNI配置文件的路径
	ConfigPath string
	// ConfigFileName 配置文件名称
	ConfigFileName string
	// NetworkSavePath 函数容器网络配置文件保存路径
	NetworkSavePath string
	// NamespaceFormat 为容器创建的Namespace格式，仅支持一个字符串类型的占位符如cqless-%s
	NamespaceFormat string

	// 以下为生成默认配置文件字段

	// NetworkName 网络名称
	NetworkName string
	// BridgeName 网桥名称
	BridgeName string
	// SubNet 网桥子网
	SubNet string
	// IfPrefix 虚拟网卡的前缀
	IfPrefix string
}

type Logger struct {
	SavePath string
	Debug    bool
}

type ProxyClient struct {
	Timeout             time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
}
