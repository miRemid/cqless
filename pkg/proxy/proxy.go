package proxy

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/logger"
	"github.com/miRemid/cqless/pkg/proxy/types"
	"github.com/miRemid/cqless/pkg/resolver"
	rtypes "github.com/miRemid/cqless/pkg/resolver/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var defaultProxy = new(Proxy)

// Proxy 代理管理中枢
// 将会以HTTP服务的形式，反向代理HTTP、Websocket和GRPC请求
type Proxy struct {
	log        zerolog.Logger
	clientPool *sync.Pool
}

func Init(opt *types.ProxyOption) error {
	return defaultProxy.Init(opt)
}

func (p *Proxy) Init(config *types.ProxyOption) error {
	p.log = log.Hook(logger.ModuleHook("proxy"))
	p.clientPool = &sync.Pool{
		New: func() any {
			return NewProxyClientFromConfig(config)
		},
	}
	return nil
}

func ReverseHandler() gin.HandlerFunc {
	return defaultProxy.ReverseHandler()
}

// ReverseHandler 反向代理
// 直接穿透给目标服务器
// 网络地址格式为: http://host_ip:port/functionName/requestURI...
func (p *Proxy) ReverseHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 提取函数名
		funcName := ctx.Param("funcName")
		if funcName == "" {
			p.log.Error().Str("uri", ctx.Request.RequestURI).Msg("no found funcname in the request uri")
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		evt := p.log.With().Str("func", funcName).Logger()

		evt.Debug().Str("raw-request", ctx.Request.URL.String()).Msg("raw proxy request")
		_items := strings.Split(ctx.Request.URL.Path, "/")[2:]
		requestURI := strings.Join(_items, "/")
		evt.Debug().Str("proxy-request-uri", requestURI).Msg("raw proxy request")

		var node = new(rtypes.Node)
		var err error
		// 指定remote node, scheme://host:port
		targetRemote := ctx.DefaultQuery("remote", "")
		if targetRemote != "" {
			parseURL, err := url.Parse(targetRemote)
			if err != nil {
				evt.Err(err).Str("target", targetRemote).Msg("parse specific remote address failed")
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}
			node.Scheme = parseURL.Scheme
			node.Host = parseURL.Host
		} else {
			node, err = p.GetRemoteNodeByName(ctx, funcName)
			if err != nil {
				evt.Err(err).Msg("get remote address faield")
				ctx.AbortWithStatus(http.StatusBadGateway)
				return
			}
		}

		// 拼接反向请求
		upstreamReq, err := p.BuildProxyRequest(ctx.Request, node, requestURI)
		if err != nil {
			evt.Err(err).Msgf("build upstream proxy request failed")
			ctx.AbortWithStatus(http.StatusBadGateway)
			return
		}

		evt.Debug().Str("endpoint", upstreamReq.URL.String()).Msg("send proxy request")
		// 发送请求
		client := p.clientPool.Get().(*http.Client)
		defer p.clientPool.Put(client)
		upstreamRes, err := client.Do(upstreamReq)
		if err != nil {
			evt.Err(err).Str("remote_addr", upstreamReq.RemoteAddr).Msg("request proxy failed")
			ctx.AbortWithStatus(http.StatusBadGateway)
			return
		}
		defer upstreamRes.Body.Close()
		// if upstreamRes.Body != nil {
		// 	defer upstreamRes.Body.Close()
		// }

		// 返回请求
		rawHeader := ctx.Writer.Header()
		copyHeaders(rawHeader, &upstreamRes.Header)
		ctx.Status(upstreamRes.StatusCode)
		// if upstreamRes.Body != nil {
		// 	n, err := io.Copy(ctx.Writer, upstreamReq.Body)
		// 	if err != nil {
		// 		evt.Err(err).Msg("rewrite proxy data failed!!!")
		// 		ctx.AbortWithStatusJSON(http.StatusInternalServerError, httputil.Response{
		// 			Code:    httputil.StatusInternalServerError,
		// 			Message: err.Error(),
		// 		})
		// 		return
		// 	}
		// 	evt.Debug().Int64("data-length", n).Msg("rewrite proxy data success")
		// }
		data, _ := io.ReadAll(upstreamRes.Body)
		ctx.Writer.Write(data)
	}
}

func (p *Proxy) GetRemoteNodeByName(ctx context.Context, funcName string) (*rtypes.Node, error) {
	return resolver.Next(ctx, funcName)
}

func (p *Proxy) BuildProxyRequest(ori *http.Request, node *rtypes.Node, requestURI string) (*http.Request, error) {
	url := node.URL()
	url.Path = requestURI
	url.RawQuery = ori.URL.RawQuery

	upstreamReq, err := http.NewRequest(ori.Method, url.String(), nil)
	if err != nil {
		return nil, err
	}
	// copy header
	upstreamReq.Header = ori.Header.Clone()
	if len(ori.Host) > 0 && upstreamReq.Header.Get("X-Forwarded-Host") == "" {
		upstreamReq.Header["X-Forwarded-Host"] = []string{ori.Host}
	}
	if upstreamReq.Header.Get("X-Forwarded-For") == "" {
		upstreamReq.Header["X-Forwarded-For"] = []string{ori.RemoteAddr}
	}

	if ori.Body != nil {
		upstreamReq.Body = ori.Body
	}

	return upstreamReq, nil
}

func NewProxyClientFromConfig(config *types.ProxyOption) *http.Client {
	return NewProxyClient(config.Timeout, config.MaxIdleConns, config.MaxIdleConnsPerHost)
}

func NewProxyClient(timeout time.Duration, maxIdleConns int, maxIdleConnsPerHost int) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 1 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          maxIdleConns,
			MaxIdleConnsPerHost:   maxIdleConnsPerHost,
			IdleConnTimeout:       120 * time.Millisecond,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// copyHeaders clones the header values from the source into the destination.
func copyHeaders(destination http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		destination[k] = vClone
	}
}
