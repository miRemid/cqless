package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
)

const (
	defaultContentType = "text/plain"
	nameExpression     = "[a-zA-Z0-9.]+"
)

var (
	nameReg         *regexp.Regexp
	proxyClientPool *sync.Pool
)

func init() {
	nameReg, _ = regexp.Compile(nameExpression)
}

func MakeProxyHandler(config *types.ProxyConfig) gin.HandlerFunc {
	return defaultGateway.MakeProxyHandler(config, cninetwork.DefaultManager)
}

func (gate *Gateway) MakeProxyHandler(config *types.ProxyConfig, cni *cninetwork.CNIManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Body != nil {
			defer ctx.Request.Body.Close()
		}

		switch ctx.Request.Method {
		case http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodGet,
			http.MethodOptions,
			http.MethodHead:
			gate.proxyRequest(ctx, cni)
		default:
			httputil.JSON(ctx, http.StatusMethodNotAllowed, httputil.Response{
				Code:    httputil.ProxyNotAllowed,
				Message: "暂未支持的请求方式: " + ctx.Request.Method,
			})
		}
	}
}

func extractFunctionProxy(req *http.Request) (string, string, error) {
	// TODO: domain support
	requestURI := req.RequestURI
	funcName := strings.Split(requestURI, "/")[0]
	if funcName == "" {
		return "", "", errors.New("未找到函数名称")
	}
	if !nameReg.MatchString(funcName) {
		return "", "", errors.New("函数名称不合法")
	}
	proxyPath := strings.ReplaceAll(requestURI, "/"+funcName, "")
	return funcName, proxyPath, nil
}

// ProxyRequest
// /funcName/
func (gate *Gateway) proxyRequest(ctx *gin.Context, cni *cninetwork.CNIManager) {
	functionName, proxyURI, err := extractFunctionProxy(ctx.Request)
	if err != nil {
		gate.log.Err(err).Msgf("解析请求 '%s' 失败", ctx.Request.RequestURI)
	}

	gate.log.Debug().Any("proxyPath", proxyURI).Str("functionName", functionName).Send()

	functionAddr, resolveErr := gate.provider.Resolve(ctx, functionName, cni)
	if resolveErr != nil {
		gate.log.Err(resolveErr).Msgf("获取目标函数 '%s' IP地址失败", functionName)
		httputil.BadRequestWithJSON(ctx, httputil.Response{
			Code:    httputil.ProxyBadRequest,
			Message: fmt.Sprintf("未找到 '%s' 函数", functionName),
		})
		return
	}
	gate.log.Debug().Str("functionAddr", functionAddr.String()).Send()

	proxyReq, err := buildProxyRequest(ctx.Request, functionAddr, proxyURI)
	if err != nil {
		gate.log.Err(err).Send()
		httputil.BadRequestWithJSON(ctx, httputil.Response{
			Code:    httputil.ProxyBadRequest,
			Message: fmt.Sprintf("获取目标函数 '%s' 地址失败", functionName),
		})
		return
	}

	if proxyReq.Body != nil {
		defer proxyReq.Body.Close()
	}

	proxyClient := proxyClientPool.Get().(*http.Client)
	defer proxyClientPool.Put(proxyClient)

	start := time.Now()
	response, err := proxyClient.Transport.RoundTrip(proxyReq.WithContext(ctx))
	seconds := time.Since(start)
	gate.log.Err(err).Msgf("请求上游函数 '%s' 失败", functionName)
	if err, ok := err.(*url.Error); ok && err.Timeout() {
		httputil.TimeoutWithJSON(ctx, httputil.Response{
			Code:    httputil.ProxyTimeout,
			Message: fmt.Sprintf("请求 '%s' 函数超时", functionName),
		})
		return
	} else if err != nil {
		httputil.BadGatewayWithJSON(ctx, httputil.Response{
			Code:    httputil.ProxyInternalServerError,
			Message: fmt.Sprintf("网络无法到达 '%s' 函数", functionName),
		})
		return
	}
	defer response.Body.Close()
	gate.log.Printf("请求函数 '%s' 共使用%f秒\n", functionName, seconds.Seconds())
	data, _ := io.ReadAll(response.Body)

	clientHeader := ctx.Writer.Header()
	copyHeaders(clientHeader, &response.Header)

	reply := httputil.Response{
		Code:    httputil.StatusOK,
		Message: "",
	}
	responseContentType := response.Header.Get("Content-Type")
	gate.log.Debug().Str("content-type", responseContentType).Str("data", string(data)).Msg("检查回复Content-Type")
	if strings.Contains(responseContentType, "json") {
		var tmpData interface{}
		if err := json.NewDecoder(bytes.NewBuffer(data)).Decode(&tmpData); err != nil {
			gate.log.Err(err).Msgf("解析函数 '%s' 失败", functionName)
			httputil.BadRequestWithJSON(ctx, httputil.Response{
				Code:    httputil.ProxyInternalServerError,
				Message: "解析函数返回数据错误",
			})
			return
		}
		reply.Data = tmpData
	} else if strings.Contains(responseContentType, "text") {
		reply.Data = string(data)
	} else {
		reply.Data = data
	}

	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(reply); err != nil {
		gate.log.Err(err).Msg("编码失败")
		httputil.BadRequestWithJSON(ctx, httputil.Response{
			Code:    httputil.ProxyInternalServerError,
			Message: "编码函数返回数据失败",
		})
		return
	}
	ctx.Writer.WriteHeader(response.StatusCode)
	ctx.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(buffer.Bytes())))
	if _, err := ctx.Writer.Write(buffer.Bytes()); err != nil {
		gate.log.Err(err).Send()
	}
}

// buildProxyRequest creates a request object for the proxy request, it will ensure that
// the original request headers are preserved as well as setting openfaas system headers
func buildProxyRequest(originalReq *http.Request, baseURL url.URL, extraPath string) (*http.Request, error) {

	host := baseURL.Host
	if baseURL.Port() == "" {
		host = baseURL.Host + ":" + types.DEFAULT_WATCHDOG_PORT
	}

	url := url.URL{
		Scheme:   baseURL.Scheme,
		Host:     host,
		Path:     extraPath,
		RawQuery: originalReq.URL.RawQuery,
	}

	upstreamReq, err := http.NewRequest(originalReq.Method, url.String(), nil)
	if err != nil {
		return nil, err
	}
	copyHeaders(upstreamReq.Header, &originalReq.Header)

	if len(originalReq.Host) > 0 && upstreamReq.Header.Get("X-Forwarded-Host") == "" {
		upstreamReq.Header["X-Forwarded-Host"] = []string{originalReq.Host}
	}
	if upstreamReq.Header.Get("X-Forwarded-For") == "" {
		upstreamReq.Header["X-Forwarded-For"] = []string{originalReq.RemoteAddr}
	}

	if originalReq.Body != nil {
		upstreamReq.Body = originalReq.Body
	}

	return upstreamReq, nil
}

// copyHeaders clones the header values from the source into the destination.
func copyHeaders(destination http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		destination[k] = vClone
	}
}

func NewProxyClientFromConfig(config *types.ProxyConfig) *http.Client {
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
