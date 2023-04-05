package gateway

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/provider"
	"github.com/miRemid/cqless/pkg/types"
)

func MakeProxyHandler(config *types.ProxyConfig) gin.HandlerFunc {
	return defaultGateway.MakeProxyHandler(config)
}

func (gate *Gateway) MakeProxyHandler(config *types.ProxyConfig) gin.HandlerFunc {
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
			provider.ProxyRequest(ctx, defaultProxyClient, gate.provider)
		default:
			httputil.JSON(ctx, http.StatusMethodNotAllowed, httputil.Response{
				Code:    httputil.ProxyNotAllowed,
				Message: "request method not allowed",
			})
		}
	}
}
