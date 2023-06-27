package cqhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/miRemid/cqless/pkg/cqhttp/types"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/logger"
	dtypes "github.com/miRemid/cqless/pkg/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var defaultCQHTTPManager *CQHTTPManager
var mutex sync.Mutex

func init() {
	defaultCQHTTPManager = new(CQHTTPManager)
}

type CQHTTPManager struct {
	mutex       sync.RWMutex
	websockets_ sync.Map
	config      *types.CQHTTPOption
	messageChan chan *CQHTTPMessage
	quickReply  chan *CQHTTPMessage
	httpClient  *http.Client

	log zerolog.Logger
}

func NewCQHTTPManager(config *types.CQHTTPOption) *CQHTTPManager {
	m := &CQHTTPManager{
		config:      config,
		websockets_: sync.Map{},
		messageChan: make(chan *CQHTTPMessage),
		quickReply:  make(chan *CQHTTPMessage),
		mutex:       sync.RWMutex{},
		httpClient:  &http.Client{},
		log:         log.Hook(logger.ModuleHook("cqhttp")),
	}
	m.log.Info().Msg("启动消息缓存队列")
	go m.processMessageQueue()
	m.log.Info().Msg("启动快速回复队列")
	go m.processQuickReployQueue()
	return m
}

func Init(opt *types.CQHTTPOption) error {
	return defaultCQHTTPManager.Init(opt)
}

func (m *CQHTTPManager) Init(opt *types.CQHTTPOption) error {
	m.config = opt
	m.websockets_ = sync.Map{}
	m.messageChan = make(chan *CQHTTPMessage)
	m.quickReply = make(chan *CQHTTPMessage)
	m.mutex = sync.RWMutex{}
	m.httpClient = &http.Client{}
	m.log = log.Hook(logger.ModuleHook("cqhttp"))
	m.log.Info().Msg("start message cache queue")
	go m.processMessageQueue()
	m.log.Info().Msg("start quick reply queue")
	go m.processQuickReployQueue()
	return nil
}

func (m *CQHTTPManager) processMessageQueue() {
	// 从消息队列中持续读取
	for {
		msg := <-m.messageChan
		log.Info().Str("message", msg.Message).Send()
		// 解析命令, 获取参数
		funcName, params, err := msg.Parser()
		if err != nil {
			log.Err(err).Send()
			continue
		}
		// 调用目标函数
		cqless_invoke_api := "http://%s:%d/function/%s"
		requestURI := fmt.Sprintf(cqless_invoke_api, "localhost", dtypes.GetConfig().Gateway.Port, funcName)
		uri, _ := url.Parse(requestURI)
		query := uri.Query()
		for _, p := range params {
			query.Add("param", p)
		}
		uri.RawQuery = query.Encode()
		var body bytes.Buffer
		body.Write(msg.Body)
		req, err := http.NewRequest(http.MethodPost, uri.String(), &body)
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
		if resp.StatusCode == 204 {
			// 和CQHTTP一致，遇到204状态码不响应
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
		// rm := &CQHTTPMessage{
		// 	BOT:         msg.BOT,
		// 	ID:          msg.ID,
		// 	MessageType: msg.MessageType,
		// }
		if data, ok := response.Data.(string); ok {
			msg.Message = data
			// rm.Message = msg
		} else if mpData, ok := response.Data.(map[string]interface{}); ok {
			if message, ok := mpData["reply"]; ok {
				if strMessage, ok := message.(string); ok {
					// rm.Message = strMessage
					msg.Message = strMessage
				}
			}
		}
		m.quickReply <- msg
	}
}

func (m *CQHTTPManager) processQuickReployQueue() {
	for {
		msg := <-m.quickReply
		tmp, ok := m.websockets_.Load(msg.BOT)
		if ok {
			wb := tmp.(*CQHTTPWebsocket)
			if err := wb.Send(msg); err != nil {
				log.Err(err).Send()
			}
		}
	}
}
