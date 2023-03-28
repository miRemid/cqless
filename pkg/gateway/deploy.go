package gateway

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/miRemid/cqless/pkg/utils"
)

func (gate *Gateway) Deploy(cni *cninetwork.CNIManager, secretMountPath string, alwaysPull bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			return
		}
		defer r.Body.Close()

		body, _ := io.ReadAll(r.Body)
		req := types.FunctionCreateRequest{}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		namespace := utils.GetRequestNamespace(req.Namespace)
		if valid, err := gate.provider.ValidNamespace(namespace); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if !valid {
			http.Error(w, types.ErrNamespaceNotFound.Error(), http.StatusBadRequest)
			return
		}
		namespaceSecretMountPath := getNamespaceSecretMountPath(secretMountPath, namespace)
		if err := validateSecrets(namespaceSecretMountPath, req.Secrets); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx := context.Background()
		fn, err := gate.provider.Deploy(ctx, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = cni.CreateCNINetwork(ctx, fn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ip, err := cni.GetIPAddress(fn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fn.IPAddress = ip
	}
}

func MakeDeployHandler(cni *cninetwork.CNIManager, secretMountPath string, alwaysPull bool) http.HandlerFunc {
	return defaultGateway.Deploy(cni, secretMountPath, alwaysPull)
}
