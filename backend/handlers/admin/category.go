package handlers

import (
	"fmt"
	"litewiki/models"
	"litewiki/services"
	"litewiki/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AdminCategory 管理员分类处理器
type AdminCategory struct{}

// List 分类列表
func (a *AdminCategory) List(c *gin.Context) {
	categories, err := services.ListCategories()
	if err != nil {
		utils.Error(c, 500, 50000, "获取分类列表失败")
		return
	}
	utils.Success(c, categories)
}

// Create 创建分类
func (a *AdminCategory) Create(c *gin.Context) {
	var req services.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40001, "请求参数错误")
		return
	}

	cat, err := services.CreateCategory(&req)
	if err != nil {
		utils.Error(c, 400, 40002, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "create_category",
		TargetType: "category",
		TargetID:   cat.ID,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, cat)
}

// Update 更新分类
func (a *AdminCategory) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的分类 ID")
		return
	}

	var req services.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40002, "请求参数错误")
		return
	}

	if err := services.UpdateCategory(id, &req); err != nil {
		utils.Error(c, 400, 40003, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "update_category",
		TargetType: "category",
		TargetID:   id,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, nil)
}

// Delete 删除分类
func (a *AdminCategory) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的分类 ID")
		return
	}

	if err := services.DeleteCategory(id); err != nil {
		utils.Error(c, 400, 40002, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "delete_category",
		TargetType: "category",
		TargetID:   id,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, nil)
}

// BatchDelete 批量删除分类
func (a *AdminCategory) BatchDelete(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40001, "请求参数错误")
		return
	}

	if len(req.IDs) == 0 {
		utils.Error(c, 400, 40002, "请选择要删除的分类")
		return
	}

	deleted, err := services.BatchDeleteCategories(req.IDs)
	if err != nil {
		utils.Error(c, 400, 40003, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "batch_delete_categories",
		TargetType: "category",
		Detail:     fmt.Sprintf("deleted %d categories", deleted),
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, gin.H{"deleted": deleted})
}

// CascadeBatchDelete 级联批量删除分类及其下所有文档
func (a *AdminCategory) CascadeBatchDelete(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40001, "请求参数错误")
		return
	}

	if len(req.IDs) == 0 {
		utils.Error(c, 400, 40002, "请选择要删除的分类")
		return
	}

	result, err := services.CascadeDeleteCategories(req.IDs)
	if err != nil {
		utils.Error(c, 400, 40003, err.Error())
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "cascade_delete_categories",
		TargetType: "category",
		Detail:     fmt.Sprintf("deleted %d categories, %d documents", result.DeletedCategories, result.DeletedDocuments),
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, result)
}

// Sort 批量排序
func (a *AdminCategory) Sort(c *gin.Context) {
	var req struct {
		Items []struct {
			ID       int64 `json:"id"`
			SortOrder int  `json:"sort_order"`
		} `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40001, "请求参数错误")
		return
	}

	// 提取 ID 列表（按 sort_order 排序）
	ids := make([]int64, len(req.Items))
	for i, item := range req.Items {
		ids[i] = item.ID
	}

	if err := services.SortCategories(ids); err != nil {
		utils.Error(c, 400, 40002, err.Error())
		return
	}

	utils.Success(c, nil)
}
