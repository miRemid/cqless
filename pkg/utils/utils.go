package utils

import (
	"net/http"

	"github.com/miRemid/cqless/pkg/types"
)

func GetRequestNamespace(namespace string) string {
	if len(namespace) > 0 {
		return namespace
	}
	return types.DEFAULT_FUNCTION_NAMESPACE
}

func GetNamespaceFromRequest(r *http.Request) string {
	q := r.URL.Query()
	namespace := q.Get("namespace")
	return GetRequestNamespace(namespace)
}
