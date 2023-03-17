package utils

import "github.com/miRemid/cqless/pkg/types"

func GetRequestNamespace(namespace string) string {
	if len(namespace) > 0 {
		return namespace
	}
	return types.DEFAULT_FUNCTION_NAMESPACE
}
