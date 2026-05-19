package services

import (
	"errors"
	"fmt"
	"litewiki/models"
)

// CreateDocumentRequest 创建文档请求
type CreateDocumentRequest struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	CategoryID *int64 `json:"category_id"`
}

// GetDocument 获取文档详情（仅已发布）
func GetDocument(id int64) (*models.Document, error) {
	doc, err := models.GetDocumentByID(id)
	if err != nil {
		return nil, errors.New("系统错误")
	}
	if doc == nil {
		return nil, errors.New("文档不存在")
	}
	if !doc.IsPublished {
		return nil, errors.New("文档未发布")
	}
	return doc, nil
}

// AdminGetDocument 管理员获取文档详情（含草稿）
func AdminGetDocument(id int64) (*models.Document, error) {
	doc, err := models.GetDocumentByID(id)
	if err != nil {
		return nil, errors.New("系统错误")
	}
	if doc == nil {
		return nil, errors.New("文档不存在")
	}
	return doc, nil
}

// ListDocuments 文档列表
func ListDocuments(page, size int, includeDraft bool) ([]models.DocumentListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return models.ListDocuments(page, size, includeDraft)
}

// ListDocumentsFiltered 带过滤的文档列表
func ListDocumentsFiltered(page, size int, includeDraft bool, filter models.DocumentFilter) ([]models.DocumentListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return models.ListDocumentsFiltered(page, size, includeDraft, filter)
}

// CreateDocument 创建文档
func CreateDocument(req *CreateDocumentRequest, authorID int64) (*models.Document, error) {
	if len(req.Title) == 0 || len(req.Title) > 200 {
		return nil, errors.New("标题需 1-200 字符")
	}

	// 验证分类是否存在
	if req.CategoryID != nil {
		cat, err := models.GetCategoryByID(*req.CategoryID)
		if err != nil || cat == nil {
			return nil, errors.New("分类不存在")
		}
	}

	doc := &models.Document{
		Title:      req.Title,
		Content:    req.Content,
		CategoryID: req.CategoryID,
		AuthorID:   authorID,
	}

	id, err := models.CreateDocument(doc)
	if err != nil {
		return nil, errors.New("创建文档失败")
	}
	doc.ID = id

	// 保存初始版本
	go models.CreateDocumentVersion(&models.DocumentVersion{
		DocumentID: id,
		Title:      req.Title,
		Content:    req.Content,
		EditorID:   authorID,
	})

	return doc, nil
}

// UpdateDocument 更新文档
func UpdateDocument(id int64, req *CreateDocumentRequest, editorID int64) error {
	if len(req.Title) == 0 || len(req.Title) > 200 {
		return errors.New("标题需 1-200 字符")
	}

	doc, err := models.GetDocumentByID(id)
	if err != nil || doc == nil {
		return errors.New("文档不存在")
	}

	// 保存版本记录（先保存旧版本）
	go models.CreateDocumentVersion(&models.DocumentVersion{
		DocumentID: id,
		Title:      doc.Title,
		Content:    doc.Content,
		EditorID:   editorID,
	})

	doc.Title = req.Title
	doc.Content = req.Content
	doc.CategoryID = req.CategoryID

	return models.UpdateDocument(doc)
}

// PublishDocument 发布/取消发布文档
func PublishDocument(id int64, published bool) error {
	doc, err := models.GetDocumentByID(id)
	if err != nil || doc == nil {
		return errors.New("文档不存在")
	}
	return models.PublishDocument(id, published)
}

// DeleteDocument 删除文档
func DeleteDocument(id int64) error {
	doc, err := models.GetDocumentByID(id)
	if err != nil || doc == nil {
		return errors.New("文档不存在")
	}
	return models.DeleteDocument(id)
}

// GetDocumentVersions 获取文档版本历史
func GetDocumentVersions(documentID int64) ([]models.DocumentVersion, error) {
	return models.ListDocumentVersions(documentID)
}

// BatchDeleteDocuments 批量删除文档
func BatchDeleteDocuments(ids []int64) (int, error) {
	if len(ids) == 0 {
		return 0, errors.New("请选择要删除的文档")
	}

	deleted := 0
	for _, id := range ids {
		doc, err := models.GetDocumentByID(id)
		if err != nil || doc == nil {
			continue // 跳过不存在的文档
		}
		if err := models.DeleteDocument(id); err != nil {
			return deleted, fmt.Errorf("删除文档 %d 失败: %w", id, err)
		}
		deleted++
	}
	return deleted, nil
}

// RollbackDocument 回滚到指定版本
func RollbackDocument(documentID, versionID int64, editorID int64) error {
	doc, err := models.GetDocumentByID(documentID)
	if err != nil || doc == nil {
		return errors.New("文档不存在")
	}

	version, err := models.GetDocumentVersionByID(versionID)
	if err != nil || version == nil {
		return errors.New("版本不存在")
	}
	if version.DocumentID != documentID {
		return errors.New("版本不属于该文档")
	}

	// 保存当前版本
	go models.CreateDocumentVersion(&models.DocumentVersion{
		DocumentID: documentID,
		Title:      doc.Title,
		Content:    doc.Content,
		EditorID:   editorID,
	})

	// 回滚
	doc.Title = version.Title
	doc.Content = version.Content
	return models.UpdateDocument(doc)
}
