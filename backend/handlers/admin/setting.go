package handlers

import (
	"litewiki/models"
	"litewiki/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminSetting 系统设置处理器
type AdminSetting struct{}

// GetSettings 获取系统设置
func (s *AdminSetting) GetSettings(c *gin.Context) {
	registerMode := models.GetRegisterMode()

	utils.Success(c, gin.H{
		"register_mode": registerMode,
	})
}

// UpdateSettingsRequest 更新系统设置请求
type UpdateSettingsRequest struct {
	RegisterMode string `json:"register_mode"`
}

// UpdateSettings 更新系统设置
func (s *AdminSetting) UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}

	// 校验 register_mode 值
	if req.RegisterMode != "" {
		if req.RegisterMode != "open" && req.RegisterMode != "invite" {
			utils.Error(c, http.StatusBadRequest, 40002, "register_mode 必须为 open 或 invite")
			return
		}
		if err := models.SetOption("register_mode", req.RegisterMode); err != nil {
			utils.Error(c, http.StatusInternalServerError, 50001, "更新注册模式失败")
			return
		}
	}

	utils.Success(c, gin.H{
		"register_mode": models.GetRegisterMode(),
	})
}
