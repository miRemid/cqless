package gateway

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/miRemid/cqless/pkg/utils"
)

func (gate *Gateway) Remove(cni *cninetwork.CNIManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			return
		}
		defer r.Body.Close()

		decoder := json.NewDecoder(r.Body)
		req := types.FunctionRemoveRequest{}
		if err := decoder.Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		namespace := utils.GetNamespaceFromRequest(r)
		if valid, err := gate.provider.ValidNamespace(namespace); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if !valid {
			http.Error(w, types.ErrNamespaceNotFound.Error(), http.StatusBadRequest)
			return
		}
		if fn, err := gate.provider.Remove(context.Background(), req); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else if err := cni.DeleteCNINetwork(context.Background(), fn); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func MakeRemoveHandler(cni *cninetwork.CNIManager) http.HandlerFunc {
	return defaultGateway.Remove(cni)
}
