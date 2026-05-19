package handlers

import (
	"litewiki/models"
	"litewiki/services"
	"litewiki/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BFFDocument BFF 文档处理器
type BFFDocument struct{}

// ListAllPublished 列出所有已发布文档（供 VitePress 构建拉取）
func (d *BFFDocument) ListAllPublished(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "100"))

	docs, total, err := services.ListDocuments(page, size, false)
	if err != nil {
		utils.Error(c, 500, 50000, "获取文档列表失败")
		return
	}

	utils.Success(c, utils.PageResult{
		List:  docs,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// GetDocument 获取已发布文档详情
func (d *BFFDocument) GetDocument(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的文档 ID")
		return
	}

	doc, err := services.GetDocument(id)
	if err != nil {
		utils.Error(c, 404, 40400, err.Error())
		return
	}

	utils.Success(c, doc)
}

// ListDocsByCategory 获取分类下的文档列表
func (d *BFFDocument) ListDocsByCategory(c *gin.Context) {
	catID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的分类 ID")
		return
	}

	docs, err := models.ListDocumentsByCategory(catID)
	if err != nil {
		utils.Error(c, 500, 50000, "获取文档列表失败")
		return
	}

	utils.Success(c, docs)
}
