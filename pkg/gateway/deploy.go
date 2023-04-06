package gateway

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/miRemid/cqless/pkg/utils"
)

func (gate *Gateway) MakeDeployHandler(cni *cninetwork.CNIManager, secretMountPath string, alwaysPull bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Body == nil {
			return
		}
		defer ctx.Request.Body.Close()

		body, _ := io.ReadAll(ctx.Request.Body)
		req := types.FunctionCreateRequest{}
		if err := json.Unmarshal(body, &req); err != nil {
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}
		namespace := utils.GetRequestNamespace(req.Namespace)
		if valid, err := gate.provider.ValidNamespace(namespace); err != nil {
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: err.Error(),
			})
			return
		} else if !valid {
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}
		namespaceSecretMountPath := getNamespaceSecretMountPath(secretMountPath, namespace)
		if err := validateSecrets(namespaceSecretMountPath, req.Secrets); err != nil {
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}
		fn, err := gate.provider.Deploy(ctx, req, cni)
		if err != nil {
			httputil.BadRequest(ctx, httputil.Response{
				Code:    httputil.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}
		httputil.OK(ctx, httputil.Response{
			Code: httputil.StatusOK,
			Data: fn,
		})
	}
}

func MakeDeployHandler(secretMountPath string, alwaysPull bool) gin.HandlerFunc {
	return defaultGateway.MakeDeployHandler(cninetwork.DefaultManager, secretMountPath, alwaysPull)
}
