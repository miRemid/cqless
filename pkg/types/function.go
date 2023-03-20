package types

// Function 用于抽象provider和cninetwork之间的沟通桥梁
type Function struct {
	ID   string // 容器ID
	PID  uint32 // 宿主机上的PID
	Name string // 函数名称

	Namespace string // 所在Namespace
	IPAddress string // CNI分配的IP地址

	EnvVars  map[string]string // 容器内的环境变量
	Metadata map[string]string // 容器的Meta数据
}
