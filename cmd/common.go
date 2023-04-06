package cmd

var (
	functionName                string
	functionNamespace           string
	httpClientReadTimeout       int
	httpClientWriteTimeout      int
	httpClientGatewayAddress    string
	httpClientGatewayConfigPath string
	verbose                     bool
)

// API
var (
	cqless_function_api = "http://%s:%d/cqless/function"
	cqless_invoke_api   = "http://%s:%d/function/%s"
)
