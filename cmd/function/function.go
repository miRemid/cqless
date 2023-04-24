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

	cqless_function_api = "http://%s:%d/cqless/function"
	cqless_invoke_api   = "http://%s:%d/function/%s"

	httpClient               *http.Client
	httpClientGatewayAddress string
	httpClientGatewayPort    int
	httpTimeout              int
)

func init() {
	functionConfigReader = viper.New()
	httpClient = http.DefaultClient
}

func Init(functionCmd *cobra.Command) {
	functionCmd.Run = functionCmd.HelpFunc()

	functionCmd.PersistentFlags().IntVarP(&httpTimeout, "timeout", "t", 30, "执行超时时间，默认30s")
	functionCmd.PersistentFlags().StringVarP(&httpClientGatewayAddress, "gateway", "g", "127.0.0.1", "网关地址，默认127.0.0.1")
	functionCmd.PersistentFlags().IntVarP(&httpClientGatewayPort, "port", "p", 8888, "网关端口，默认8888")
	functionCmd.PersistentFlags().StringVar(&functionNamespace, "namespace", types.DEFAULT_FUNCTION_NAMESPACE, "函数所在命名空间(Docker无需关心)")
	functionCmd.PersistentFlags().StringVarP(&functionConfigPath, "config", "c", "", "函数部署配置文件路径，默认为空")

	functionCmd.AddCommand(deployCmd)
	functionCmd.AddCommand(inspectCmd)
	functionCmd.AddCommand(invokeCmd)
	functionCmd.AddCommand(rmCmd)
}
