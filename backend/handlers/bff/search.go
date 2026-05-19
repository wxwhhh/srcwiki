package handlers

import (
	"litewiki/services"
	"litewiki/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BFFSearch BFF 搜索处理器
type BFFSearch struct{}

// Search 全文搜索
func (s *BFFSearch) Search(c *gin.Context) {
	query := c.Query("q")
	if len(query) == 0 {
		utils.Error(c, 400, 40001, "搜索关键词不能为空")
		return
	}
	if len(query) > 100 {
		utils.Error(c, 400, 40002, "搜索关键词过长")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	results, total, err := services.Search(query, page, size)
	if err != nil {
		utils.Error(c, 500, 50000, "搜索失败")
		return
	}

	utils.Success(c, utils.PageResult{
		List:  results,
		Total: total,
		Page:  page,
		Size:  size,
	})
}
