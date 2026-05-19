// Package handlers 提供 HTTP 请求处理器
package handlers

import (
	"litewiki/models"
	"litewiki/services"
	"litewiki/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BFFAuth BFF 认证处理器
type BFFAuth struct {
	JWTSecret string
}

// Register 邀请码注册
func (a *BFFAuth) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}

	// 验证验证码
	if !VerifyCaptcha(req.CaptchaID, req.CaptchaCode) {
		utils.Error(c, http.StatusBadRequest, 40003, "验证码错误或已过期")
		return
	}

	user, err := services.Register(&req)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, 40002, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     user.ID,
		Username:   user.Username,
		Action:     "register",
		TargetType: "user",
		TargetID:   user.ID,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}

// Login 用户登录
func (a *BFFAuth) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}

	// 验证验证码
	if !VerifyCaptcha(req.CaptchaID, req.CaptchaCode) {
		utils.Error(c, http.StatusBadRequest, 40003, "验证码错误或已过期")
		return
	}

	token, user, err := services.Login(&req, a.JWTSecret)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, 40102, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     user.ID,
		Username:   user.Username,
		Action:     "login",
		TargetType: "user",
		TargetID:   user.ID,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	// 设置 HttpOnly Cookie
	c.SetCookie("token", token, 86400, "/", "", false, true)

	utils.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}

// Logout 用户登出
func (a *BFFAuth) Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	utils.Success(c, nil)
}

// GetMe 获取当前用户信息
func (a *BFFAuth) GetMe(c *gin.Context) {
	userID := c.GetInt64("user_id")
	user, err := models.GetUserByID(userID)
	if err != nil || user == nil {
		utils.Error(c, http.StatusNotFound, 40400, "用户不存在")
		return
	}

	utils.Success(c, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"role":       user.Role,
		"status":     user.Status,
		"created_at": user.CreatedAt,
	})
}

// UpdateProfile 修改个人信息（目前仅用户名，暂不支持）
func (a *BFFAuth) UpdateProfile(c *gin.Context) {
	utils.Success(c, nil)
}

// ChangePassword 修改密码
func (a *BFFAuth) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}

	userID := c.GetInt64("user_id")
	if err := services.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		utils.Error(c, http.StatusBadRequest, 40003, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     userID,
		Username:   c.GetString("username"),
		Action:     "change_password",
		TargetType: "user",
		TargetID:   userID,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, nil)
}
