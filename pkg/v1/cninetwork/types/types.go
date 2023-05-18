package types

type NetworkOption struct {
	// BinaryPath CNI插件的绝对路径
	BinaryPath string `yaml:"binary_path" mapstructure:"binary_path"`
	// ConfigPath 存放CNI配置文件的路径
	ConfigPath string `yaml:"config_path" mapstructure:"config_path"`
	// ConfigFileName 配置文件名称
	ConfigFileName string `yaml:"config_file_name" mapstructure:"config_file_name"`
	// NetworkSavePath 函数容器网络配置文件保存路径
	NetworkSavePath string `yaml:"network_save_path" mapstructure:"network_save_path"`
	// NamespaceFormat 为容器创建的Namespace格式，仅支持一个字符串类型的占位符如cqless-%s
	NamespaceFormat string `yaml:"namespace_format" mapstructure:"namespace_format"`

	// 以下为生成默认配置文件字段

	// NetworkName 网络名称
	NetworkName string `yaml:"network_name" mapstructure:"network_name"`
	// BridgeName 网桥名称
	BridgeName string `yaml:"bridge_name" mapstructure:"bridge_name"`
	// SubNet 网桥子网
	SubNet string `yaml:"subnet" mapstructure:"subnet"`
	// IfPrefix 虚拟网卡的前缀
	IfPrefix string `yaml:"if_prefix" mapstructure:"if_prefix"`
}
