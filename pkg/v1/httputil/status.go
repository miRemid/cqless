package httputil

const (
	StatusOK = 0

	StatusBad = 100 + iota
	StatusBadRequest
	StatusInternalServerError

	Proxy = 200 + iota
	ProxyNotAllowed
	ProxyNotFound

	ProxyBadRequest          // 请求参数错误
	ProxyInternalServerError // 函数内存错误，通常为非超时造成的错误
	ProxyTimeout             // 超时错误
)

const (
	ErrBadRequestParams = "解析请求数据失败"
)
