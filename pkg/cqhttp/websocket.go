package cqhttp

import "github.com/gorilla/websocket"

type Manager struct {
	Connections map[uint]*websocket.Conn
}
