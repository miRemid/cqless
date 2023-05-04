package cqhttp

import (
	"fmt"
	"strings"
)

type CQHTTPMessage struct {
	BOT         uint   // 机器人ID
	ID          uint   // 发送者ID
	GroupID     uint   // 群组ID，可能
	MessageType string // Private or Group
	Message     string
	Body        []byte
}

func (m *CQHTTPMessage) Parser() (string, []string, error) {
	// TODO: 添加支持，目前直接处理!!开头的情况
	if !strings.Contains(m.Message, "!!") {
		return "", []string{}, fmt.Errorf("非命令过滤")
	}
	seps := strings.Split(m.Message[2:], " ")
	cmd := seps[0]
	return cmd, seps[1:], nil
}
