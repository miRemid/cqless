package cqhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var defaultCQHTTPManager *CQHTTPManager
var mutex sync.Mutex

func GetDefaultCQHTTPManager() *CQHTTPManager {
	if defaultCQHTTPManager == nil {
		mutex.Lock()
		if defaultCQHTTPManager == nil {
			defaultCQHTTPManager = NewCQHTTPManager(types.GetConfig())
		}
		mutex.Unlock()
	}
	return defaultCQHTTPManager
}

type CQHTTPManager struct {
	mutex       sync.RWMutex
	websockets_ sync.Map
	config      *types.CQHTTPConfig
	messageChan chan *CQHTTPMessage
	quickReply  chan *CQHTTPMessage
	httpClient  *http.Client
}

func NewCQHTTPManager(config *types.CQLessConfig) *CQHTTPManager {
	m := &CQHTTPManager{
		config:      config.CQHTTP,
		websockets_: sync.Map{},
		messageChan: make(chan *CQHTTPMessage),
		quickReply:  make(chan *CQHTTPMessage),
		mutex:       sync.RWMutex{},
		httpClient:  &http.Client{},
	}
	log.Info().Msg("启动消息缓存队列")
	go m.processMessageQueue()
	log.Info().Msg("启动快速回复队列")
	go m.processQuickReployQueue()
	return m
}

func (m *CQHTTPManager) processMessageQueue() {
	// 从消息队列中持续读取
	for {
		msg := <-m.messageChan
		log.Info().Msg(msg.Message)
		// 解析命令, 获取参数
		funcName, err := msg.Parser()
		if err != nil {
			log.Err(err).Send()
			continue
		}
		// 调用目标函数
		cqless_invoke_api := "http://%s:%d/function/%s"
		requestURI := fmt.Sprintf(cqless_invoke_api, "localhost", types.GetConfig().Gateway.Port, funcName)
		var body bytes.Buffer
		body.Write(msg.Body)
		req, err := http.NewRequest(http.MethodPost, requestURI, &body)
		if err != nil {
			log.Err(errors.Wrap(err, "解析失败或过滤")).Send()
			continue
		}
		resp, err := m.httpClient.Do(req)
		if err != nil {
			log.Err(errors.Wrapf(err, "调用 ‘%s’ 函数失败", funcName)).Send()
			resp.Body.Close()
			continue
		}
		if resp.StatusCode != 200 {
			log.Error().Msg("调用远程函数失败，请检查网关进程日志")
			resp.Body.Close()
			continue
		}
		var response httputil.Response
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Err(errors.Wrap(err, "parse response failed")).Send()
			resp.Body.Close()
			continue
		}
		// 检查返回值
		if response.Code != httputil.StatusOK {
			log.Error().Msgf("调用远程函数失败: %s", response.Message)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()
		rm := &CQHTTPMessage{
			ID:          msg.ID,
			MessageType: msg.MessageType,
		}
		if msg, ok := response.Data.(string); ok {
			rm.Message = msg
		} else if msg, ok := response.Data.(map[string]interface{}); ok {
			rm.Message = msg["reply"].(string)
		}
		m.quickReply <- rm
	}
}

func (m *CQHTTPManager) processQuickReployQueue() {
	for {
		msg := <-m.quickReply
		tmp, ok := m.websockets_.Load(msg.ID)
		if ok {
			wb := tmp.(*CQHTTPWebsocket)
			if err := wb.Send(msg); err != nil {
				log.Err(err).Send()
			}
		}
	}
}
