package httputil

const (
	StatusOK = 0

	StatusBad = 100 + iota
	StatusBadRequest
	StatusInternalServerError

	Proxy = 200 + iota
	ProxyNotAllowed
	ProxyNotFound
	ProxyBadRequest
	ProxyInternalServerError
)
