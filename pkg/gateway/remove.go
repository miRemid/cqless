package gateway

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/miRemid/cqless/pkg/utils"
)

func (gate *Gateway) MakeRemoveHandler(cni *cninetwork.CNIManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Body == nil {
			return
		}
		defer ctx.Request.Body.Close()

		req := types.FunctionRemoveRequest{}
		if err := ctx.BindJSON(&req); err != nil {
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}
		namespace := utils.GetNamespaceFromRequest(ctx.Request)
		if valid, err := gate.provider.ValidNamespace(namespace); err != nil {
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: err.Error(),
			})
			return
		} else if !valid {
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: types.ErrNamespaceNotFound.Error(),
			})
			return
		}
		if fn, err := gate.provider.Remove(context.Background(), req, cni); err != nil {
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: types.ErrNamespaceNotFound.Error(),
			})
		} else {
			httputil.OK(ctx, httputil.Response{
				Code: httputil.StatusOK,
				Data: fn,
			})
		}
	}
}

func MakeRemoveHandler() gin.HandlerFunc {
	return defaultGateway.MakeRemoveHandler(cninetwork.DefaultManager)
}
