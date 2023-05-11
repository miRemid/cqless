package cqhttp

import (
	"bytes"
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type CQHTTPWebsocket struct {
	id   uint
	conn *websocket.Conn

	log zerolog.Logger
}

func (c CQHTTPWebsocket) Run(e *zerolog.Event, level zerolog.Level, message string) {
	e.Uint("bot_id", c.id)
}

func (c *CQHTTPWebsocket) Close() error {
	err := c.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *CQHTTPWebsocket) Listen(ch chan *CQHTTPMessage) error {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}
		c.log.Debug().Msg(string(message))
		postType, _ := jsonparser.GetString(message, "post_type")
		// TODO: 后续添加其他PostType支持，目前仅支持Message
		if postType != "message" {
			continue
		}
		var msg = new(CQHTTPMessage)
		id, _ := jsonparser.GetInt(message, "self_id")
		uid, _ := jsonparser.GetInt(message, "user_id")
		_msg, _ := jsonparser.GetString(message, "raw_message")
		_msgType, _ := jsonparser.GetString(message, "message_type")
		if _msgType == "group" {
			gid, _ := jsonparser.GetInt(message, "group_id")
			msg.GroupID = uint(gid)
		}
		msg.BOT = uint(id)
		msg.ID = uint(uid)
		msg.Body = message
		msg.Message = _msg
		msg.MessageType = _msgType
		ch <- msg
	}
}

func (c *CQHTTPWebsocket) Send(msg *CQHTTPMessage) error {
	var buffer bytes.Buffer
	var reply = map[string]interface{}{}
	if msg.MessageType == "private" {
		reply["action"] = "send_private_msg"
		reply["params"] = map[string]interface{}{
			"user_id": msg.ID,
			"message": msg.Message,
		}
	} else {
		reply["action"] = "send_group_msg"
		reply["params"] = map[string]interface{}{
			"message":  msg.Message,
			"group_id": msg.GroupID,
		}
	}
	if err := json.NewEncoder(&buffer).Encode(reply); err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
}
