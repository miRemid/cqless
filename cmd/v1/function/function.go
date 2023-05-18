package function

import (
	"net/http"

	"github.com/miRemid/cqless/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	functionName         string // 用于删除、获取、调用
	functionNamespace    string // 用于删除、获取、调用
	functionConfigPath   string // 用于创建函数以及构建函数的配置文件路径
	functionConfigReader *viper.Viper

	cqless_invoke_api = "http://%s:%d/%s"

	httpClient          *http.Client
	httpClientProxyPort int
	timeout             int

	grpcServerAddress string
)

func init() {
	functionConfigReader = viper.New()
	httpClient = &http.Client{}
}

func Init(functionCmd *cobra.Command) {
	functionCmd.Run = functionCmd.HelpFunc()

	functionCmd.PersistentFlags().IntVarP(&timeout, "timeout", "", 30, "execute timeout，default 30s")
	functionCmd.PersistentFlags().StringVarP(&grpcServerAddress, "gateway", "g", "127.0.0.1:5565", "cqless gateway server，default 127.0.0.1:5565")
	functionCmd.PersistentFlags().StringVar(&functionNamespace, "namespace", types.DEFAULT_FUNCTION_NAMESPACE, "function namespace(useless for Docker)")
	functionCmd.PersistentFlags().StringVarP(&functionConfigPath, "config", "c", "", "function config path，default empty")

	functionCmd.AddCommand(deployCmd)
	functionCmd.AddCommand(inspectCmd)
	functionCmd.AddCommand(invokeCmd)
	functionCmd.AddCommand(rmCmd)
}
