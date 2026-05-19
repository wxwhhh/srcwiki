// Package handlers 提供 HTTP 请求处理器
package handlers

import (
	"litewiki/models"
	"litewiki/services"
	"litewiki/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AdminUser 管理员用户处理器
type AdminUser struct{}

// List 用户列表
func (a *AdminUser) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	users, total, err := services.ListUsers(page, size)
	if err != nil {
		utils.Error(c, 500, 50000, "获取用户列表失败")
		return
	}

	utils.Success(c, utils.PageResult{
		List:  users,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// Delete 删除用户（不能删除自己）
func (a *AdminUser) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的用户 ID")
		return
	}

	operatorID := c.GetInt64("user_id")
	if err := services.DeleteUser(id, operatorID); err != nil {
		utils.Error(c, 400, 40002, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     operatorID,
		Username:   c.GetString("username"),
		Action:     "delete_user",
		TargetType: "user",
		TargetID:   id,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, nil)
}

// UpdateRole 修改用户角色
func (a *AdminUser) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的用户 ID")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40002, "请求参数错误")
		return
	}

	if err := services.UpdateUserRole(id, req.Role); err != nil {
		utils.Error(c, 400, 40003, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "update_user_role",
		TargetType: "user",
		TargetID:   id,
		Detail:     "role -> " + req.Role,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, nil)
}

// UpdateStatus 修改用户状态
func (a *AdminUser) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的用户 ID")
		return
	}

	// 防止管理员禁用自己
	operatorID := c.GetInt64("user_id")
	if id == operatorID {
		utils.Error(c, 400, 40004, "不能修改自己的状态")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40002, "请求参数错误")
		return
	}

	if err := services.UpdateUserStatus(id, req.Status); err != nil {
		utils.Error(c, 400, 40003, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "update_user_status",
		TargetType: "user",
		TargetID:   id,
		Detail:     "status -> " + req.Status,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, nil)
}
