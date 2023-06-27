package gateway

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/miRemid/cqless/pkg/provider"
	"github.com/miRemid/cqless/pkg/resolver"
	"github.com/miRemid/cqless/pkg/types"
)

func (gate *Gateway) MakeDeployHandler(cni *cninetwork.CNIManager, secretMountPath string, alwaysPull bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		evt := gate.log.With().Str("action", "deploy").Logger()
		if ctx.Request.Body == nil {
			return
		}
		defer ctx.Request.Body.Close()

		body, _ := io.ReadAll(ctx.Request.Body)
		req := types.FunctionCreateRequest{}
		if err := json.Unmarshal(body, &req); err != nil {
			evt.Err(err).Msg(httputil.ErrBadRequestParams)
			httputil.BadRequest(ctx)
			return
		}
		namespace := GetRequestNamespace(req.Namespace)
		if valid, err := provider.ValidNamespace(namespace); err != nil || !valid {
			evt := evt.Error()
			if err != nil {
				evt.Err(err)
			}
			evt.Msg("校验namespace失败")
			httputil.BadRequest(ctx)
			return
		}
		namespaceSecretMountPath := getNamespaceSecretMountPath(secretMountPath, namespace)
		if err := validateSecrets(namespaceSecretMountPath, req.Secrets); err != nil {
			evt.Err(err).Msg("校验secretsMountPath失败")
			httputil.BadRequest(ctx)
			return
		}
		fn, err := provider.Deploy(ctx, req, cni)
		if err != nil {
			evt.Err(err).Msgf("创建函数 '%s' 失败", req.Name)
			httputil.OKWithJSON(ctx, httputil.Response{
				Code:    httputil.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}
		if err := resolver.Register(ctx, fn.Name, fn.Node()); err != nil {
			gate.log.Err(err).Msgf("DNS插入失败")
			httputil.OKWithJSON(ctx, httputil.Response{
				Code:    httputil.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}
		gate.log.Info().Str("函数名", fn.Name).Msg("创建函数成功")
		httputil.OKWithJSON(ctx, httputil.Response{
			Code: httputil.StatusOK,
			Data: fn,
		})
	}
}

func MakeDeployHandler(secretMountPath string, alwaysPull bool) gin.HandlerFunc {
	return defaultGateway.MakeDeployHandler(cninetwork.DefaultManager, secretMountPath, alwaysPull)
}
