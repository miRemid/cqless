package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		used := time.Since(start)
		log.Info().Str("path", ctx.Request.RequestURI).Dur("used", used).Str("method", ctx.Request.Method).Send()
	}
}
