package services

import (
	"errors"
	"fmt"
	"litewiki/models"
)

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID *int64 `json:"parent_id"`
	SortOrder int  `json:"sort_order"`
}

// GetCategoryTree 获取分类树
func GetCategoryTree() ([]models.CategoryTreeNode, error) {
	return models.GetCategoryTree()
}

// ListCategories 平铺分类列表
func ListCategories() ([]models.Category, error) {
	return models.ListCategories()
}

// CreateCategory 创建分类
func CreateCategory(req *CreateCategoryRequest) (*models.Category, error) {
	if len(req.Name) == 0 || len(req.Name) > 64 {
		return nil, errors.New("分类名需 1-64 字符")
	}

	// 验证父分类是否存在
	if req.ParentID != nil {
		parent, err := models.GetCategoryByID(*req.ParentID)
		if err != nil || parent == nil {
			return nil, errors.New("父分类不存在")
		}
	}

	cat := &models.Category{
		ParentID:  req.ParentID,
		Name:      req.Name,
		SortOrder: req.SortOrder,
	}

	id, err := models.CreateCategory(cat)
	if err != nil {
		return nil, errors.New("创建分类失败")
	}
	cat.ID = id

	return cat, nil
}

// UpdateCategory 更新分类
func UpdateCategory(id int64, req *CreateCategoryRequest) error {
	if len(req.Name) == 0 || len(req.Name) > 64 {
		return errors.New("分类名需 1-64 字符")
	}

	existing, err := models.GetCategoryByID(id)
	if err != nil || existing == nil {
		return errors.New("分类不存在")
	}

	// 不能将自己设为自己的父分类
	if req.ParentID != nil && *req.ParentID == id {
		return errors.New("不能将分类设为自己的子分类")
	}

	existing.Name = req.Name
	existing.ParentID = req.ParentID
	existing.SortOrder = req.SortOrder

	return models.UpdateCategory(existing)
}

// DeleteCategory 删除分类
func DeleteCategory(id int64) error {
	existing, err := models.GetCategoryByID(id)
	if err != nil || existing == nil {
		return errors.New("分类不存在")
	}
	return models.DeleteCategory(id)
}

// SortCategories 批量排序分类
func SortCategories(ids []int64) error {
	if len(ids) == 0 {
		return errors.New("排序列表不能为空")
	}
	return models.UpdateCategorySort(ids)
}

// BatchDeleteCategories 批量删除分类
func BatchDeleteCategories(ids []int64) (int, error) {
	if len(ids) == 0 {
		return 0, errors.New("请选择要删除的分类")
	}

	deleted := 0
	for _, id := range ids {
		existing, err := models.GetCategoryByID(id)
		if err != nil || existing == nil {
			continue // 跳过不存在的分类
		}
		if err := models.DeleteCategory(id); err != nil {
			return deleted, fmt.Errorf("删除分类 %d 失败: %w", id, err)
		}
		deleted++
	}
	return deleted, nil
}

// CascadeDeleteCategoriesResult 级联删除结果
type CascadeDeleteCategoriesResult struct {
	DeletedCategories int `json:"deleted_categories"`
	DeletedDocuments  int `json:"deleted_documents"`
}

// CascadeDeleteCategories 级联批量删除分类及其下所有文档
func CascadeDeleteCategories(ids []int64) (*CascadeDeleteCategoriesResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("请选择要删除的分类")
	}

	// 验证所有分类是否存在
	for _, id := range ids {
		existing, err := models.GetCategoryByID(id)
		if err != nil || existing == nil {
			return nil, fmt.Errorf("分类 %d 不存在", id)
		}
	}

	deletedCats, deletedDocs, err := models.CascadeDeleteCategories(ids)
	if err != nil {
		return nil, fmt.Errorf("级联删除失败: %w", err)
	}

	return &CascadeDeleteCategoriesResult{
		DeletedCategories: deletedCats,
		DeletedDocuments:  deletedDocs,
	}, nil
}

// FindCategoryByNameAndParent 根据名称和父分类查找分类
func FindCategoryByNameAndParent(name string, parentID *int64) (*models.Category, error) {
	return models.GetCategoryByNameAndParent(name, parentID)
}
