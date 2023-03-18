package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/miRemid/cqless/pkg/cninetwork"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/miRemid/cqless/pkg/utils"
)

var (
	defaultProvider *Provider
)

func init() {
	defaultProvider = new(Provider)
}

func InitProvider() {

}

type ProviderPluginInterface interface {
	ValidNamespace(string) (bool, error)
	Deploy(ctx context.Context, req types.FunctionDeployRequest, cni *cninetwork.CNIManager, namespace string, alwaysPull bool) error

	Close()
}

type Provider struct {
	plugin ProviderPluginInterface
}

func (p *Provider) Deploy(cni *cninetwork.CNIManager, secretMountPath string, alwaysPull bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			return
		}
		defer r.Body.Close()

		body, _ := io.ReadAll(r.Body)
		req := types.FunctionDeployRequest{}
		if err := json.Unmarshal(body, &req); err != nil {
			return
		}
		namespace := utils.GetRequestNamespace(req.Namespace)
		if valid, err := p.plugin.ValidNamespace(namespace); err != nil {
			return
		} else if !valid {
			return
		}
		namespaceSecretMountPath := getNamespaceSecretMountPath(secretMountPath, namespace)
		if err := validateSecrets(namespaceSecretMountPath, req.Secrets); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx := context.Background()
		if err := p.plugin.Deploy(ctx, req, cni, namespaceSecretMountPath, alwaysPull); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func MakeDeployHandler(cni *cninetwork.CNIManager, secretMountPath string, alwaysPull bool) http.HandlerFunc {
	return defaultProvider.Deploy(cni, secretMountPath, alwaysPull)
}
