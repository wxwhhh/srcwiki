package handlers

import (
	"litewiki/models"
	"litewiki/utils"

	"github.com/gin-gonic/gin"
)

// BFFCredits BFF 致谢处理器
type BFFCredits struct{}

// List 获取所有致谢（公开接口，需登录）
func (b *BFFCredits) List(c *gin.Context) {
	credits, err := models.GetAllCredits()
	if err != nil {
		utils.Error(c, 500, 50000, "获取致谢列表失败")
		return
	}
	utils.Success(c, credits)
}
