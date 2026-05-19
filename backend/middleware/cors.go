package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// 允许的 Origin 白名单
var allowedOrigins = []string{
	"http://src-wiki.com",
	"http://45.207.211.144",
	"http://localhost:8080",
	"http://127.0.0.1:8080",
}

// CORS 跨域配置中间件（白名单模式）
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" && isAllowedOrigin(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isAllowedOrigin 检查 origin 是否在白名单中
func isAllowedOrigin(origin string) bool {
	for _, allowed := range allowedOrigins {
		if strings.EqualFold(origin, allowed) {
			return true
		}
	}
	return false
}
