package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (gate *Gateway) MakeInspectHandler(cni *cninetwork.CNIManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Body == nil {
			return
		}
		defer ctx.Request.Body.Close()
		// check params
		var req types.FunctionInspectRequest
		if err := ctx.Bind(&req); err != nil {
			log.Err(err).Send()
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.ProxyBadRequest,
				Message: errors.Wrapf(err, "binding failed").Error(),
			})
			return
		}
		fns, err := gate.provider.Inspect(ctx, req, cni)
		if err != nil {
			log.Err(err).Send()
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.ProxyBadRequest,
				Message: errors.Wrapf(err, "get functions failed").Error(),
			})
			return
		}
		httputil.OK(ctx, httputil.Response{
			Code: httputil.StatusOK,
			Data: fns,
		})

	}
}

func MakeInspectHandler() gin.HandlerFunc {
	return defaultGateway.MakeInspectHandler(cninetwork.DefaultManager)
}
