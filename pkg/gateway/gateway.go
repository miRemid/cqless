package gateway

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/miRemid/cqless/pkg/gateway/types"
	"github.com/miRemid/cqless/pkg/logger"
	dtypes "github.com/miRemid/cqless/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	defaultGateway *Gateway
)

func init() {
	defaultGateway = new(Gateway)
}

func Init(config *types.GatewayOption) error {

	return defaultGateway.Init(config)
}

type Gateway struct {
	log zerolog.Logger
}

func (gate *Gateway) Init(config *types.GatewayOption) error {
	gate.log = log.Hook(logger.ModuleHook("gateway"))
	// https://github.com/rfyiamcool/notes/blob/main/golang_net_http_optimize.md
	return nil
}

func getNamespaceSecretMountPath(userSecretPath string, namespace string) string {
	return path.Join(userSecretPath, namespace)
}
func validateSecrets(secretMountPath string, secrets []string) error {
	for _, secret := range secrets {
		if _, err := os.Stat(path.Join(secretMountPath, secret)); err != nil {
			return fmt.Errorf("unable to find secret: %s", secret)
		}
	}
	return nil
}
func GetRequestNamespace(namespace string) string {
	if len(namespace) > 0 {
		return namespace
	}
	return dtypes.DEFAULT_FUNCTION_NAMESPACE
}

func GetNamespaceFromRequest(r *http.Request) string {
	q := r.URL.Query()
	namespace := q.Get("namespace")
	return GetRequestNamespace(namespace)
}
