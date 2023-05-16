package gateway

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/provider"
	"github.com/miRemid/cqless/pkg/resolver"
	"github.com/miRemid/cqless/pkg/types"
)

func (gate *Gateway) MakeRemoveHandler(cni *cninetwork.CNIManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Body == nil {
			return
		}
		defer ctx.Request.Body.Close()

		req := types.FunctionRemoveRequest{}
		if err := ctx.BindJSON(&req); err != nil {
			gate.log.Err(err).Msg(httputil.ErrBadRequestParams)
			httputil.BadRequest(ctx)
			return
		}
		namespace := GetNamespaceFromRequest(ctx.Request)
		if valid, err := provider.ValidNamespace(namespace); err != nil || !valid {
			evt := gate.log.Error()
			if err != nil {
				evt.Err(err)
			}
			evt.Msg("校验namespace失败")
			httputil.BadRequest(ctx)
			return
		}
		if err := provider.Remove(ctx, req, cni); err != nil {
			gate.log.Err(err).Msg("remove function failed")
			httputil.OKWithJSON(ctx, httputil.Response{
				Code:    httputil.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}

		if err := resolver.UnRegisterFunc(ctx, req.FunctionName); err != nil {
			gate.log.Err(err).Msg("remove dns failed")
			httputil.OKWithJSON(ctx, httputil.Response{
				Code:    httputil.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}

		httputil.OKWithJSON(ctx, httputil.Response{
			Code:    httputil.StatusOK,
			Message: fmt.Sprintf("函数 `%s` 已被成功删除", req.FunctionName),
		})
	}
}

func MakeRemoveHandler() gin.HandlerFunc {
	return defaultGateway.MakeRemoveHandler(cninetwork.DefaultManager)
}
