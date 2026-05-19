package handlers

import (
	"litewiki/services"
	"litewiki/utils"

	"github.com/gin-gonic/gin"
)

// BFFTree BFF 分类树处理器
type BFFTree struct{}

// GetTree 获取完整分类树
func (t *BFFTree) GetTree(c *gin.Context) {
	tree, err := services.GetCategoryTree()
	if err != nil {
		utils.Error(c, 500, 50000, "获取分类树失败")
		return
	}
	utils.Success(c, tree)
}
