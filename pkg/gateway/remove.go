package gateway

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/miRemid/cqless/pkg/utils"
)

func (gate *Gateway) Remove(cni *cninetwork.CNIManager) gin.HandlerFunc {
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
		if fn, err := gate.provider.Remove(context.Background(), req); err != nil {
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: types.ErrNamespaceNotFound.Error(),
			})
			return
		} else if err := cni.DeleteCNINetwork(context.Background(), fn); err != nil {
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: types.ErrNamespaceNotFound.Error(),
			})
			return
		}
	}
}

func MakeRemoveHandler(cni *cninetwork.CNIManager) gin.HandlerFunc {
	return defaultGateway.Remove(cni)
}
