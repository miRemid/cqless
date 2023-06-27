package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miRemid/cqless/pkg/gateway/types"
	"github.com/miRemid/cqless/pkg/httputil"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

func RateLimit(rateConfig *types.RateLimitOption) gin.HandlerFunc {
	log.Info().Float64("rateLimit.limit", rateConfig.Limit).Int("rateLimit.Burst", rateConfig.Burst).Msg("已启动令牌桶功能")
	limit := rate.NewLimiter(rate.Limit(rateConfig.Limit), rateConfig.Burst)
	return func(ctx *gin.Context) {
		if limit.Allow() {
			ctx.Next()
		} else {
			log.Debug().Msg("过多请求，拒绝调用")
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, httputil.Response{})
		}
	}
}
