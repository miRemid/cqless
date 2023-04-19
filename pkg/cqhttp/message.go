package cqhttp

import (
	"fmt"
	"strings"
)

type CQHTTPMessage struct {
	ID          uint   // 发送者ID
	GroupID     uint   // 群组ID，可能
	MessageType string // Private or Group
	Message     string
	Body        []byte
}

func (m *CQHTTPMessage) Parser() (string, error) {
	// TODO: 添加支持，目前直接处理!!开头的情况
	if !strings.Contains(m.Message, "!!") {
		return "", fmt.Errorf("非命令过滤")
	}
	return m.Message[2:], nil
}
