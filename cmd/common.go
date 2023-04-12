package cmd

var (
	functionName       string // 用于删除、获取、调用
	functionNamespace  string // 用于删除、获取、调用
	functionConfigPath string // 用于创建函数以及构建函数的配置文件路径

	httpClientReadTimeout       int
	httpClientWriteTimeout      int
	httpClientGatewayAddress    string
	httpClientGatewayConfigPath string
)

// API
var (
	cqless_function_api = "http://%s:%d/cqless/function"
	cqless_invoke_api   = "http://%s:%d/function/%s"
)
