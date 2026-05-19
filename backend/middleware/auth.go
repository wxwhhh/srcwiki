// Package middleware 提供 Gin 中间件
package middleware

import (
	"litewiki/utils"

	"github.com/gin-gonic/gin"
)

// Auth JWT 鉴权中间件：从 HttpOnly Cookie 或 Authorization Header 读取 token
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// 1. 优先从 Authorization Header 读取 Bearer Token
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}

		// 2. 如果 Header 没有，从 Cookie 读取
		if token == "" {
			cookieToken, err := c.Cookie("token")
			if err == nil && cookieToken != "" {
				token = cookieToken
			}
		}

		if token == "" {
			utils.Error(c, 401, 40100, "未登录")
			c.Abort()
			return
		}

		// 3. 验证 JWT
		claims, err := utils.ParseJWT(token, jwtSecret)
		if err != nil {
			utils.Error(c, 401, 40101, "Token 无效或已过期")
			c.Abort()
			return
		}

		// 4. 注入用户信息到 Context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}
