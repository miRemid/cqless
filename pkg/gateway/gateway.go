package gateway

import (
	"fmt"
	"os"
	"path"

	"github.com/miRemid/cqless/pkg/provider"
	"github.com/miRemid/cqless/pkg/types"
)

var (
	defaultGateway *Gateway
)

func init() {
	defaultGateway = new(Gateway)
}

func Init(config *types.CQLessConfig) error {
	return defaultGateway.Init(config)
}

type Gateway struct {
	provider provider.ProviderPluginInterface
}

func (gate *Gateway) Init(config *types.CQLessConfig) error {
	if err := gate.provider.Init(config); err != nil {
		return err
	}
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
