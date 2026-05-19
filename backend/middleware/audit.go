package middleware

import (
	"litewiki/models"

	"github.com/gin-gonic/gin"
)

// AuditLog 审计日志中间件：记录操作日志
func AuditLog(action, targetType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // 先执行 handler

		// 获取用户信息（可能不存在，如未登录接口）
		var userID int64
		var username string
		if uid, exists := c.Get("user_id"); exists {
			userID = uid.(int64)
		}
		if uname, exists := c.Get("username"); exists {
			username = uname.(string)
		}

		// 异步写入审计日志（不阻塞响应）
		go models.InsertAuditLog(&models.AuditLog{
			UserID:     userID,
			Username:   username,
			Action:     action,
			TargetType: targetType,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
		})
	}
}
