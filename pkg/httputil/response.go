package httputil

type Response struct {
	Code    int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}
