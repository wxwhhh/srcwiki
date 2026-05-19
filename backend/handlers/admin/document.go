package handlers

import (
	"fmt"
	"litewiki/models"
	"litewiki/services"
	"litewiki/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AdminDocument 管理员文档处理器
type AdminDocument struct{}

// List 文档列表（含草稿）
func (a *AdminDocument) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	var filter models.DocumentFilter
	if catIDStr := c.Query("category_id"); catIDStr != "" {
		if catID, err := strconv.ParseInt(catIDStr, 10, 64); err == nil {
			filter.CategoryID = &catID
		}
	}
	filter.Status = c.Query("status")

	docs, total, err := services.ListDocumentsFiltered(page, size, true, filter)
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

// Get 获取文档详情（含草稿）
func (a *AdminDocument) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的文档 ID")
		return
	}

	doc, err := services.AdminGetDocument(id)
	if err != nil {
		utils.Error(c, 404, 40400, err.Error())
		return
	}

	utils.Success(c, doc)
}

// Create 创建文档
func (a *AdminDocument) Create(c *gin.Context) {
	var req services.CreateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40001, "请求参数错误")
		return
	}

	doc, err := services.CreateDocument(&req, c.GetInt64("user_id"))
	if err != nil {
		utils.Error(c, 400, 40002, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "create_document",
		TargetType: "document",
		TargetID:   doc.ID,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, doc)
}

// Update 编辑文档
func (a *AdminDocument) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的文档 ID")
		return
	}

	var req services.CreateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40002, "请求参数错误")
		return
	}

	if err := services.UpdateDocument(id, &req, c.GetInt64("user_id")); err != nil {
		utils.Error(c, 400, 40003, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "update_document",
		TargetType: "document",
		TargetID:   id,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, nil)
}

// Delete 删除文档
func (a *AdminDocument) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的文档 ID")
		return
	}

	if err := services.DeleteDocument(id); err != nil {
		utils.Error(c, 400, 40002, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "delete_document",
		TargetType: "document",
		TargetID:   id,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, nil)
}

// BatchDelete 批量删除文档
func (a *AdminDocument) BatchDelete(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40001, "请求参数错误")
		return
	}

	if len(req.IDs) == 0 {
		utils.Error(c, 400, 40002, "请选择要删除的文档")
		return
	}

	deleted, err := services.BatchDeleteDocuments(req.IDs)
	if err != nil {
		utils.Error(c, 400, 40003, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "batch_delete_documents",
		TargetType: "document",
		Detail:     fmt.Sprintf("deleted %d documents", deleted),
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, gin.H{"deleted": deleted})
}

// Publish 发布/取消发布文档
func (a *AdminDocument) Publish(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的文档 ID")
		return
	}

	var req struct {
		Published bool `json:"is_published"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40002, "请求参数错误")
		return
	}

	if err := services.PublishDocument(id, req.Published); err != nil {
		utils.Error(c, 400, 40003, err.Error())
		return
	}

	// 审计日志
	action := "unpublish_document"
	if req.Published {
		action = "publish_document"
	}
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     action,
		TargetType: "document",
		TargetID:   id,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, nil)
}

// GetVersions 获取文档版本历史
func (a *AdminDocument) GetVersions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的文档 ID")
		return
	}

	versions, err := services.GetDocumentVersions(id)
	if err != nil {
		utils.Error(c, 500, 50000, "获取版本历史失败")
		return
	}

	utils.Success(c, versions)
}

// Rollback 回滚到指定版本
func (a *AdminDocument) Rollback(c *gin.Context) {
	docID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的文档 ID")
		return
	}

	vid, err := strconv.ParseInt(c.Param("vid"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40002, "无效的版本 ID")
		return
	}

	if err := services.RollbackDocument(docID, vid, c.GetInt64("user_id")); err != nil {
		utils.Error(c, 400, 40003, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "rollback_document",
		TargetType: "document",
		TargetID:   docID,
		Detail:     "rollback to version " + strconv.FormatInt(vid, 10),
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, nil)
}
