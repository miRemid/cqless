package cqhttp

import (
	"net/http"

	"github.com/buger/jsonparser"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (c *CQHTTPManager) WebsocketHandler(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		c.log.Error().Err(err).Send()
		return
	}
	_, message, err := conn.ReadMessage()
	if err != nil {
		c.log.Error().Err(err).Send()
		return
	}
	post_type, err := jsonparser.GetString(message, "post_type")
	if err != nil {
		c.log.Error().Err(err).Send()
		conn.Close()
		return
	}
	if post_type != "meta_event" {
		c.log.Error().Msg("wrong post_type, websocket connect's post_type must be 'meta_event'")
		conn.Close()
		return
	}
	metaEventType, err := jsonparser.GetString(message, "meta_event_type")
	if err != nil {
		c.log.Error().Err(err).Send()
		conn.Close()
		return
	}
	if metaEventType != "lifecycle" {
		c.log.Error().Msg("wrong meta_event_type, websocket connect's meta_event_type must be 'lifecycle'")
		conn.Close()
		return
	}
	id, err := jsonparser.GetInt(message, "self_id")
	if err != nil {
		c.log.Error().Err(err).Send()
		conn.Close()
		return
	}
	wb := &CQHTTPWebsocket{
		id:   uint(id),
		conn: conn,
	}
	wb.log = c.log.Hook(wb)
	c.websockets_.Store(uint(id), wb)
	go func(cb *CQHTTPWebsocket) {
		defer cb.Close()
		defer func() {
			if err := recover(); err != nil {
				c.log.Error().Any("panic", err).Send()
			}
		}()
		if err := cb.Listen(c.messageChan); err != nil {
			panic(err)
		}
	}(wb)
	c.log.Info().Msgf("已与ID=%d的机器人建立Websocket连接", id)
}

// func (c *CQHTTPManager) SendMessageHandler(ctx *gin.Context) {
// 	// 将请求内容，转发至Websocket中
// 	// 内容格式务必填写BOT的QQ号
// 	var req = new(CQHTTPMessageRequest)
// 	if err := ctx.Bind(req); err != nil {
// 		httputil.BadRequest(ctx, httputil.Response{
// 			Code:    httputil.StatusBadRequest,
// 			Message: "发送数据格式错误",
// 		})
// 		return
// 	}

// }
