package types

// Container 用于抽象provider和cninetwork之间的沟通桥梁
type Container struct {
	Name string
	ID   string
	PID  uint32

	NetNamespace string
}
