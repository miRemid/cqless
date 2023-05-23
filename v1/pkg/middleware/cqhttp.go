package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// https://docs.go-cqhttp.org/reference/#%E4%B8%8A%E6%8A%A5%E7%AD%BE%E5%90%8D
// CqhttpAuth: Body签名校验
func CqhttpAuthHTTP(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		hSignature := c.GetHeader("X-Signature")
		if hSignature != "" {
			if secret == "" {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			// 读取 request body
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			mac := hmac.New(sha1.New, []byte(secret))
			if _, err := mac.Write(body); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			if "sha1="+hex.EncodeToString(mac.Sum(nil)) != hSignature {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}
		c.Next()
	}
}
