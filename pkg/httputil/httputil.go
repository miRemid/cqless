package httputil

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func JSON(ctx *gin.Context, statusCode int, reply any) {
	ctx.JSON(statusCode, reply)
}

func OK(ctx *gin.Context, reply any) {
	JSON(ctx, http.StatusOK, reply)
}

func BadRequest(ctx *gin.Context, reply any) {
	JSON(ctx, http.StatusBadRequest, reply)
}
