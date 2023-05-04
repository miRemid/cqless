package gateway

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/miRemid/cqless/pkg/gateway/resolver"
	"github.com/miRemid/cqless/pkg/provider"
	"github.com/miRemid/cqless/pkg/provider/docker"
	"github.com/miRemid/cqless/pkg/types"
	"github.com/rs/zerolog/log"
)

var (
	defaultGateway     *Gateway
	defaultProxyClient *http.Client
)

func init() {
	defaultGateway = new(Gateway)
	defaultProxyClient = http.DefaultClient
}

func Init(config *types.CQLessConfig) error {
	defaultProxyClient = provider.NewProxyClientFromConfig(config.Proxy)
	return defaultGateway.Init(config.Gateway)
}

type Gateway struct {
	provider provider.ProviderPluginInterface

	dns *resolver.Resolver
}

func (gate *Gateway) Init(config *types.GatewayConfig) error {

	providerType := strings.ToUpper(config.Provider)
	switch providerType {
	case "DOCKER":
		gate.provider = docker.NewProvider()
	default:
		providerType = "DOCKER"
		gate.provider = docker.NewProvider()
	}
	log.Info().Msgf("正在使用: '%s' 作为Provider", providerType)
	if err := gate.provider.Init(config); err != nil {
		return err
	}
	gate.dns = resolver.NewResolverFromConfig(config.Resolver)
	log.Info().Msgf("正在使用：'%s' 作为Resolver", config.Resolver.Type)
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
