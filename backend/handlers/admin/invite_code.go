package handlers

import (
	"fmt"
	"litewiki/models"
	"litewiki/services"
	"litewiki/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AdminInviteCode 管理员邀请码处理器
type AdminInviteCode struct{}

// BatchCreate 批量生成邀请码
func (a *AdminInviteCode) BatchCreate(c *gin.Context) {
	var req services.BatchCreateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40001, "请求参数错误：count 范围 1-500，role 必填")
		return
	}

	codes, err := services.BatchGenerateInviteCodes(&req, c.GetInt64("user_id"))
	if err != nil {
		utils.Error(c, 400, 40002, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "batch_create_invite_codes",
		TargetType: "invite_code",
		Detail:     fmt.Sprintf("批量生成 %d 个邀请码，角色: %s", len(codes), req.Role),
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, gin.H{
		"codes": codes,
		"count": len(codes),
	})
}

// List 邀请码列表
func (a *AdminInviteCode) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	codes, total, err := services.ListInviteCodes(page, size)
	if err != nil {
		utils.Error(c, 500, 50000, "获取邀请码列表失败")
		return
	}

	utils.Success(c, utils.PageResult{
		List:  codes,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// Create 生成邀请码
func (a *AdminInviteCode) Create(c *gin.Context) {
	var req services.CreateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40001, "请求参数错误")
		return
	}

	code, err := services.GenerateInviteCode(&req, c.GetInt64("user_id"))
	if err != nil {
		utils.Error(c, 400, 40002, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "create_invite_code",
		TargetType: "invite_code",
		TargetID:   code.ID,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, code)
}

// Delete 作废邀请码
func (a *AdminInviteCode) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的邀请码 ID")
		return
	}

	if err := services.DeleteInviteCode(id); err != nil {
		utils.Error(c, 500, 50000, "作废邀请码失败")
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "delete_invite_code",
		TargetType: "invite_code",
		TargetID:   id,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, nil)
}
