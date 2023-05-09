package gateway

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/miRemid/cqless/pkg/utils"
	"github.com/rs/zerolog/log"
)

func (gate *Gateway) MakeRemoveHandler(cni *cninetwork.CNIManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Body == nil {
			return
		}
		defer ctx.Request.Body.Close()

		req := types.FunctionRemoveRequest{}
		if err := ctx.BindJSON(&req); err != nil {
			log.Err(err).Msg(httputil.ErrBadRequestParams)
			httputil.BadRequest(ctx)
			return
		}
		namespace := utils.GetNamespaceFromRequest(ctx.Request)
		if valid, err := gate.provider.ValidNamespace(namespace); err != nil || !valid {
			evt := log.Error()
			if err != nil {
				evt.Err(err)
			}
			evt.Msg("校验namespace失败")
			httputil.BadRequest(ctx)
			return
		}
		if err := gate.provider.Remove(context.Background(), req, cni); err != nil {
			log.Err(err).Send()
			httputil.OKWithJSON(ctx, httputil.Response{
				Code: httputil.StatusBadRequest,
			})
		} else {
			httputil.OKWithJSON(ctx, httputil.Response{
				Code:    httputil.StatusOK,
				Message: fmt.Sprintf("函数 `%s` 已被成功删除", req.FunctionName),
			})
		}
	}
}

func MakeRemoveHandler() gin.HandlerFunc {
	return defaultGateway.MakeRemoveHandler(cninetwork.DefaultManager)
}
