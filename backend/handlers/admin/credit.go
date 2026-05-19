package handlers

import (
	"litewiki/models"
	"litewiki/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AdminCredit 管理员致谢处理器
type AdminCredit struct{}

// CreateCreditRequest 创建/更新致谢请求
type CreateCreditRequest struct {
	Name        string `json:"name" binding:"required"`
	URL         string `json:"url" binding:"required"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
	License     string `json:"license"`
	Stars       string `json:"stars"`
	SortOrder   int    `json:"sort_order"`
}

// List 致谢列表
func (a *AdminCredit) List(c *gin.Context) {
	credits, err := models.GetAllCredits()
	if err != nil {
		utils.Error(c, 500, 50000, "获取致谢列表失败")
		return
	}
	utils.Success(c, credits)
}

// Create 创建致谢
func (a *AdminCredit) Create(c *gin.Context) {
	var req CreateCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40001, "请求参数错误")
		return
	}

	credit := &models.Credit{
		Name:        req.Name,
		URL:         req.URL,
		Description: req.Description,
		IconURL:     req.IconURL,
		License:     req.License,
		Stars:       req.Stars,
		SortOrder:   req.SortOrder,
	}

	id, err := models.CreateCredit(credit)
	if err != nil {
		utils.Error(c, 500, 50001, "创建致谢失败")
		return
	}
	credit.ID = id

	utils.Success(c, credit)
}

// Update 更新致谢
func (a *AdminCredit) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的致谢 ID")
		return
	}

	existing, err := models.GetCreditByID(id)
	if err != nil {
		utils.Error(c, 500, 50000, "查询致谢失败")
		return
	}
	if existing == nil {
		utils.Error(c, 404, 40400, "致谢不存在")
		return
	}

	var req CreateCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40002, "请求参数错误")
		return
	}

	existing.Name = req.Name
	existing.URL = req.URL
	existing.Description = req.Description
	existing.IconURL = req.IconURL
	existing.License = req.License
	existing.Stars = req.Stars
	existing.SortOrder = req.SortOrder

	if err := models.UpdateCredit(existing); err != nil {
		utils.Error(c, 500, 50001, "更新致谢失败")
		return
	}

	utils.Success(c, existing)
}

// Delete 删除致谢
func (a *AdminCredit) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的致谢 ID")
		return
	}

	existing, err := models.GetCreditByID(id)
	if err != nil {
		utils.Error(c, 500, 50000, "查询致谢失败")
		return
	}
	if existing == nil {
		utils.Error(c, 404, 40400, "致谢不存在")
		return
	}

	if err := models.DeleteCredit(id); err != nil {
		utils.Error(c, 500, 50001, "删除致谢失败")
		return
	}

	utils.Success(c, nil)
}
