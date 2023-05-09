package gateway

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/miRemid/cqless/pkg/utils"
	"github.com/rs/zerolog/log"
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
			log.Err(err).Msg(httputil.ErrBadRequestParams)
			httputil.BadRequest(ctx)
			return
		}
		namespace := utils.GetRequestNamespace(req.Namespace)
		if valid, err := gate.provider.ValidNamespace(namespace); err != nil || !valid {
			evt := log.Error()
			if err != nil {
				evt.Err(err)
			}
			evt.Msg("校验namespace失败")
			httputil.BadRequest(ctx)
			return
		}
		namespaceSecretMountPath := getNamespaceSecretMountPath(secretMountPath, namespace)
		if err := validateSecrets(namespaceSecretMountPath, req.Secrets); err != nil {
			log.Err(err).Msg("校验secretsMountPath失败")
			httputil.BadRequest(ctx)
			return
		}
		fn, err := gate.provider.Deploy(ctx, req, cni)
		if err != nil {
			log.Err(err).Msgf("创建函数 '%s' 失败", req.Name)
			httputil.OKWithJSON(ctx, httputil.Response{
				Code:    httputil.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}
		log.Info().Str("函数名", fn.Name).Msg("创建函数成功")
		httputil.OKWithJSON(ctx, httputil.Response{
			Code: httputil.StatusOK,
			Data: fn,
		})
	}
}

func MakeDeployHandler(secretMountPath string, alwaysPull bool) gin.HandlerFunc {
	return defaultGateway.MakeDeployHandler(cninetwork.DefaultManager, secretMountPath, alwaysPull)
}
