package provider

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/miRemid/cqless/pkg/types"

	"github.com/rs/zerolog/log"
)

const (
	watchdogPort       = "8080"
	defaultContentType = "text/plain"
)

func (p *Provider) MakeProxyHandler(config *types.ProxyClient) http.HandlerFunc {
	proxyClient := NewProxyClientFromConfig(config)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}
		switch r.Method {
		case http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodGet,
			http.MethodOptions,
			http.MethodHead:
			proxyRequest(w, r, proxyClient, p.plugin)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// NewProxyClientFromConfig creates a new http.Client designed for proxying requests and enforcing
// certain minimum configuration values.
func NewProxyClientFromConfig(config *types.ProxyClient) *http.Client {
	return NewProxyClient(config.Timeout, config.MaxIdleConns, config.MaxIdleConnsPerHost)
}

// NewProxyClient creates a new http.Client designed for proxying requests, this is exposed as a
// convenience method for internal or advanced uses. Most people should use NewProxyClientFromConfig.
func NewProxyClient(timeout time.Duration, maxIdleConns int, maxIdleConnsPerHost int) *http.Client {
	return &http.Client{
		// these Transport values ensure that the http Client will eventually timeout and prevents
		// infinite retries. The default http.Client configure these timeouts.  The specific
		// values tuned via performance testing/benchmarking
		//
		// Additional context can be found at
		// - https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
		// - https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
		//
		// Additionally, these overrides for the default client enable re-use of connections and prevent
		// CoreDNS from rate limiting under high traffic
		//
		// See also two similar projects where this value was updated:
		// https://github.com/prometheus/prometheus/pull/3592
		// https://github.com/minio/minio/pull/5860
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

// proxyRequest handles the actual resolution of and then request to the function service.
func proxyRequest(w http.ResponseWriter, originalReq *http.Request, proxyClient *http.Client, plugin ProviderPluginInterface) {
	ctx := originalReq.Context()

	pathVars := mux.Vars(originalReq)
	functionName := pathVars["name"]
	if functionName == "" {
		http.Error(w, "Provide function name in the request path", http.StatusBadRequest)
		return
	}

	functionAddr, resolveErr := plugin.Resolve(ctx, functionName)
	if resolveErr != nil {
		// TODO: Should record the 404/not found error in Prometheus.
		http.Error(w, "No endpoints available for: "+functionName, http.StatusServiceUnavailable)
		return
	}

	proxyReq, err := buildProxyRequest(originalReq, functionAddr, pathVars["params"])
	if err != nil {
		http.Error(w, "Failed to resolve service: %s."+functionName, http.StatusServiceUnavailable)
		return
	}

	if proxyReq.Body != nil {
		defer proxyReq.Body.Close()
	}

	start := time.Now()
	response, err := proxyClient.Do(proxyReq.WithContext(ctx))
	seconds := time.Since(start)

	if err != nil {
		http.Error(w, "Can't reach service for: %s."+functionName, http.StatusInternalServerError)
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
		io.Copy(w, response.Body)
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
