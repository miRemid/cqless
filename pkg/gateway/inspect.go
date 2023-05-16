package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/provider"
	"github.com/miRemid/cqless/pkg/types"
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
			gate.log.Err(err).Send()
			httputil.BadRequest(ctx)
			return
		}
		fns, err := provider.Inspect(ctx, req, cni)
		if err != nil {
			gate.log.Err(err).Msgf("获取函数 '%s' 信息失败", req.FunctionName)
			httputil.OKWithJSON(ctx, httputil.Response{
				Code:    httputil.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}
		httputil.OKWithJSON(ctx, httputil.Response{
			Code: httputil.StatusOK,
			Data: fns,
		})

	}
}

func MakeInspectHandler() gin.HandlerFunc {
	return defaultGateway.MakeInspectHandler(cninetwork.DefaultManager)
}
