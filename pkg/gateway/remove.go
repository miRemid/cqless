package gateway

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/miRemid/cqless/pkg/utils"
	"github.com/pkg/errors"
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
			log.Err(err).Send()
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: errors.Wrap(err, "remove failed, please check gateway log").Error(),
			})
			return
		}
		namespace := utils.GetNamespaceFromRequest(ctx.Request)
		if valid, err := gate.provider.ValidNamespace(namespace); err != nil {
			log.Err(err).Send()
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: errors.Wrap(err, "remove failed, please check gateway log").Error(),
			})
			return
		} else if !valid {
			err = types.ErrNamespaceNotFound
			log.Err(err).Send()
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: errors.Wrap(err, "remove failed, please check gateway log").Error(),
			})
			return
		}
		if fn, err := gate.provider.Remove(context.Background(), req, cni); err != nil {
			log.Err(err).Send()
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: errors.Wrap(errors.Errorf("`%s` not found", req.FunctionName), "remove failed, please check gateway log").Error(),
			})
		} else {
			httputil.OK(ctx, httputil.Response{
				Code:    httputil.StatusOK,
				Data:    fn,
				Message: fmt.Sprintf("remove success, `%s` has been removed", req.FunctionName),
			})
		}
	}
}

func MakeRemoveHandler() gin.HandlerFunc {
	return defaultGateway.MakeRemoveHandler(cninetwork.DefaultManager)
}
