package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/rs/zerolog/log"
)

const (
	WatchdogPort       = "8080"
	defaultContentType = "text/plain"
	nameExpression     = "[a-zA-Z0-9.]+"
)

var (
	nameReg *regexp.Regexp
)

func init() {
	nameReg, _ = regexp.Compile(nameExpression)
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

func ProxyRequest(ctx *gin.Context, proxyClient *http.Client, plugin ProviderPluginInterface, cni *cninetwork.CNIManager) {
	// w, originalReq := ctx.Writer, ctx.Request
	functionName := ctx.Param("name")
	if functionName == "" {
		httputil.BadRequest(ctx, httputil.Response{
			Code:    httputil.ProxyNotFound,
			Message: "未找到函数名称",
		})
		return
	}
	log.Debug().Any("params", ctx.Params).Str("functionName", functionName).Send()
	if !nameReg.MatchString(functionName) {
		httputil.BadRequest(ctx, httputil.Response{
			Code:    httputil.ProxyBadRequest,
			Message: "函数名称不合法",
		})
		return
	}

	functionAddr, resolveErr := plugin.Resolve(ctx, functionName, cni)
	if resolveErr != nil {
		log.Err(resolveErr).Send()
		httputil.BadRequest(ctx, httputil.Response{
			Code:    httputil.ProxyBadRequest,
			Message: fmt.Sprintf("未找到 '%s' 函数", functionName),
		})
		return
	}
	log.Debug().Any("functionAddr", functionAddr).Send()

	proxyReq, err := buildProxyRequest(ctx.Request, functionAddr, ctx.Param("params"))
	if err != nil {
		log.Err(err).Send()
		httputil.BadRequest(ctx, httputil.Response{
			Code:    httputil.ProxyBadRequest,
			Message: fmt.Sprintf("获取目标函数 '%s' 地址失败", functionName),
		})
		return
	}

	if proxyReq.Body != nil {
		defer proxyReq.Body.Close()
	}

	start := time.Now()
	response, err := proxyClient.Do(proxyReq.WithContext(ctx))
	seconds := time.Since(start)
	if err != nil {
		log.Err(err).Send()
		httputil.BadRequest(ctx, httputil.Response{
			Code:    httputil.ProxyInternalServerError,
			Message: fmt.Sprintf("网络无法到达 '%s' 函数", functionName),
		})
		return
	}
	defer response.Body.Close()
	log.Printf("%s took %f seconds\n", functionName, seconds.Seconds())
	data, _ := io.ReadAll(response.Body)

	clientHeader := ctx.Writer.Header()
	copyHeaders(clientHeader, &response.Header)
	ctx.Writer.WriteHeader(response.StatusCode)

	reply := httputil.Response{
		Code:    httputil.StatusOK,
		Message: "",
	}
	responseContentType := response.Header.Get("Content-Type")
	if strings.Contains(responseContentType, "json") {
		var tmpData = make(map[string]interface{})
		if err := json.NewDecoder(bytes.NewBuffer(data)).Decode(&tmpData); err != nil {
			log.Err(err).Send()
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.ProxyInternalServerError,
				Message: "解析函数返回数据错误",
			})
		}
		reply.Data = tmpData
	} else if strings.Contains(responseContentType, "text") {
		reply.Data = string(data)
	} else {
		reply.Data = data
	}

	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(reply); err != nil {
		log.Err(err).Send()
		httputil.BadRequest(ctx, httputil.Response{
			Code:    httputil.ProxyInternalServerError,
			Message: fmt.Sprintf("Can't reach service for: %s.", functionName),
		})
		return
	}
	ctx.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(buffer.Bytes())))
	if _, err := ctx.Writer.Write(buffer.Bytes()); err != nil {
		log.Err(err).Send()
	}
}

// buildProxyRequest creates a request object for the proxy request, it will ensure that
// the original request headers are preserved as well as setting openfaas system headers
func buildProxyRequest(originalReq *http.Request, baseURL url.URL, extraPath string) (*http.Request, error) {

	host := baseURL.Host
	if baseURL.Port() == "" {
		host = baseURL.Host + ":" + WatchdogPort
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
