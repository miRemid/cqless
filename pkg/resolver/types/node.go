package types

import (
	"context"
	"fmt"
	"net/url"
)

type UnRegisterCallback func(ctx context.Context, funcName string, node *Node) error

type Node struct {
	Scheme   string
	Host     string // host or host:port
	FuncName string
	Metadata any
}

func NewNode(scheme, host, funcName string, metadata any) *Node {
	return &Node{
		Scheme:   scheme,
		Host:     host,
		FuncName: funcName,
		Metadata: metadata,
	}
}

func (n Node) String() string {
	return n.URL().String()
}

func (n Node) URL() *url.URL {
	return &url.URL{
		Scheme: n.Scheme,
		Host:   n.Host,
	}
}

func (n Node) GetValueKey() string {
	return fmt.Sprintf("%s/%s", n.FuncName, n.String())
}
