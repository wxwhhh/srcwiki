package handlers

import (
	"litewiki/database"
	"litewiki/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// BFFHome BFF 首页统计处理器
type BFFHome struct{}

// HomeStatsResponse 首页统计数据
type HomeStatsResponse struct {
	TotalDocs   int64              `json:"total_docs"`
	TotalCats   int64              `json:"total_cats"`
	RecentDocs  []RecentDocItem    `json:"recent_docs"`
}

// RecentDocItem 最近更新文档项
type RecentDocItem struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	CategoryName string    `json:"category_name"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetHomeStats 获取首页统计信息
func (h *BFFHome) GetHomeStats(c *gin.Context) {
	// 文档总数（已发布）
	var totalDocs int64
	err := database.DB.QueryRow("SELECT COUNT(*) FROM documents WHERE is_published = 1").Scan(&totalDocs)
	if err != nil {
		utils.Error(c, 500, 50000, "获取文档总数失败")
		return
	}

	// 分类总数
	var totalCats int64
	err = database.DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&totalCats)
	if err != nil {
		utils.Error(c, 500, 50000, "获取分类总数失败")
		return
	}

	// 最近更新的 10 篇已发布文档
	rows, err := database.DB.Query(`
		SELECT d.id, d.title, COALESCE(c.name, '未分类') as category_name, d.updated_at
		FROM documents d
		LEFT JOIN categories c ON d.category_id = c.id
		WHERE d.is_published = 1
		ORDER BY d.updated_at DESC
		LIMIT 10
	`)
	if err != nil {
		utils.Error(c, 500, 50000, "获取最近文档失败")
		return
	}
	defer rows.Close()

	var recentDocs []RecentDocItem
	for rows.Next() {
		var doc RecentDocItem
		if err := rows.Scan(&doc.ID, &doc.Title, &doc.CategoryName, &doc.UpdatedAt); err != nil {
			continue
		}
		recentDocs = append(recentDocs, doc)
	}

	if recentDocs == nil {
		recentDocs = []RecentDocItem{}
	}

	utils.Success(c, HomeStatsResponse{
		TotalDocs:  totalDocs,
		TotalCats:  totalCats,
		RecentDocs: recentDocs,
	})
}
