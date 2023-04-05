package provider

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/rs/zerolog/log"
)

const (
	watchdogPort       = "8080"
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

func ProxyRequest(ctx *gin.Context, proxyClient *http.Client, plugin ProviderPluginInterface) {
	w, originalReq := ctx.Writer, ctx.Request
	functionName := ctx.Param("name")
	if functionName == "" {
		httputil.BadRequest(ctx, httputil.Response{
			Code:    httputil.ProxyNotFound,
			Message: "Provider function name not in the request path",
		})
		return
	}
	log.Debug().Any("params", ctx.Params).Str("functionName", functionName).Send()
	if !nameReg.MatchString(functionName) {
		httputil.BadRequest(ctx, httputil.Response{
			Code:    httputil.ProxyBadRequest,
			Message: "Provider function name not valid",
		})
		return
	}

	functionAddr, resolveErr := plugin.Resolve(ctx, functionName)
	if resolveErr != nil {
		httputil.BadRequest(ctx, httputil.Response{
			Code:    httputil.ProxyBadRequest,
			Message: "No endpoints available for: " + functionName,
		})
		return
	}
	log.Debug().Any("functionAddr", functionAddr).Send()

	proxyReq, err := buildProxyRequest(ctx.Request, functionAddr, ctx.Param("params"))
	if err != nil {
		httputil.BadRequest(ctx, httputil.Response{
			Code:    httputil.ProxyBadRequest,
			Message: fmt.Sprintf("Failed to resolve service: %s.", functionName),
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
		httputil.BadRequest(ctx, httputil.Response{
			Code:    httputil.ProxyInternalServerError,
			Message: fmt.Sprintf("Can't reach service for: %s.", functionName),
		})
		return
	}

	if response.Body != nil {
		defer response.Body.Close()
	}

	log.Printf("%s took %f seconds\n", functionName, seconds.Seconds())

	clientHeader := w.Header()
	copyHeaders(clientHeader, &response.Header)
	w.Header().Set("Content-Type", getContentType(originalReq.Header, response.Header))

	w.WriteHeader(response.StatusCode)
	if response.Body != nil {
		if _, err := io.Copy(w, response.Body); err != nil {
			log.Error().Msg(fmt.Sprintf("write proxy failed: %v", err))
		}
	}
}

// buildProxyRequest creates a request object for the proxy request, it will ensure that
// the original request headers are preserved as well as setting openfaas system headers
func buildProxyRequest(originalReq *http.Request, baseURL url.URL, extraPath string) (*http.Request, error) {

	host := baseURL.Host
	if baseURL.Port() == "" {
		host = baseURL.Host + ":" + watchdogPort
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

// getContentType resolves the correct Content-Type for a proxied function.
func getContentType(request http.Header, proxyResponse http.Header) (headerContentType string) {
	responseHeader := proxyResponse.Get("Content-Type")
	requestHeader := request.Get("Content-Type")

	if len(responseHeader) > 0 {
		headerContentType = responseHeader
	} else if len(requestHeader) > 0 {
		headerContentType = requestHeader
	} else {
		headerContentType = defaultContentType
	}

	return headerContentType
}
