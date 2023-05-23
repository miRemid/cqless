package httputil

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func JSON(ctx *gin.Context, statusCode int, reply any) {
	ctx.JSON(statusCode, reply)
}

func OK(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}

func OKWithJSON(ctx *gin.Context, reply any) {
	JSON(ctx, http.StatusOK, reply)
}

func BadRequest(ctx *gin.Context) {
	ctx.Status(http.StatusBadRequest)
}

func BadRequestWithJSON(ctx *gin.Context, data any) {
	JSON(ctx, http.StatusBadRequest, data)
}

func BadGateway(ctx *gin.Context) {
	ctx.Status(http.StatusBadGateway)
}

func BadGatewayWithJSON(ctx *gin.Context, data any) {
	JSON(ctx, http.StatusBadGateway, data)
}

func InternalError(ctx *gin.Context) {
	ctx.Status(http.StatusInternalServerError)
}

func InternalErrorWithJSON(ctx *gin.Context, data any) {
	JSON(ctx, http.StatusInternalServerError, data)
}

func Timeout(ctx *gin.Context) {
	ctx.Status(http.StatusGatewayTimeout)
}

func TimeoutWithJSON(ctx *gin.Context, data any) {
	JSON(ctx, http.StatusGatewayTimeout, data)
}
