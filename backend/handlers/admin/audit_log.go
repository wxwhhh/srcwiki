package handlers

import (
	"litewiki/models"
	"litewiki/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AdminAuditLog 管理员审计日志处理器
type AdminAuditLog struct{}

// LoginStats 登录统计
func (a *AdminAuditLog) LoginStats(c *gin.Context) {
	stats, err := models.GetLoginStats()
	if err != nil {
		utils.Error(c, 500, 50000, "获取登录统计失败")
		return
	}
	utils.Success(c, stats)
}

// List 审计日志查询
func (a *AdminAuditLog) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	userID, _ := strconv.ParseInt(c.DefaultQuery("user_id", "0"), 10, 64)
	action := c.Query("action")

	logs, total, err := models.ListAuditLogs(page, size, userID, action)
	if err != nil {
		utils.Error(c, 500, 50000, "获取审计日志失败")
		return
	}

	utils.Success(c, utils.PageResult{
		List:  logs,
		Total: total,
		Page:  page,
		Size:  size,
	})
}
