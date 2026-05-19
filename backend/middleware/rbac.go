package middleware

import (
	"litewiki/utils"

	"github.com/gin-gonic/gin"
)

// RequireRole 角色权限中间件：检查用户是否具有指定角色之一
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			utils.Error(c, 403, 40300, "无权限")
			c.Abort()
			return
		}

		role := userRole.(string)
		for _, allowed := range roles {
			if role == allowed {
				c.Next()
				return
			}
		}

		utils.Error(c, 403, 40301, "权限不足")
		c.Abort()
	}
}
