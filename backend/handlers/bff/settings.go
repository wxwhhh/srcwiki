package handlers

import (
	"litewiki/models"
	"litewiki/utils"

	"github.com/gin-gonic/gin"
)

// BFFSettings BFF 系统设置处理器（公开接口）
type BFFSettings struct{}

// GetPublicSettings 获取公开系统设置（无需鉴权）
func (s *BFFSettings) GetPublicSettings(c *gin.Context) {
	registerMode := models.GetRegisterMode()

	utils.Success(c, gin.H{
		"register_mode": registerMode,
	})
}
