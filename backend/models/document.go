package models

import (
	"database/sql"
	"litewiki/database"
	"litewiki/utils"
	"time"
)

// Document 文档模型
type Document struct {
	ID          int64     `json:"id"`
	CategoryID  *int64    `json:"category_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	AuthorID    int64     `json:"author_id"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DocumentListItem 文档列表项（不含内容）
type DocumentListItem struct {
	ID          int64     `json:"id"`
	CategoryID  *int64    `json:"category_id"`
	Title       string    `json:"title"`
	AuthorID    int64     `json:"author_id"`
	AuthorName  string    `json:"author_name"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateDocument 创建文档
func CreateDocument(d *Document) (int64, error) {
	result, err := database.DB.Exec(
		"INSERT INTO documents (category_id, title, content, author_id, is_published) VALUES (?, ?, ?, ?, ?)",
		d.CategoryID, d.Title, d.Content, d.AuthorID, d.IsPublished,
	)
	if err != nil {
		return 0, err
	}

	id, _ := result.LastInsertId()

	// 同步到 FTS 索引（中文分词后索引）
	_, err = database.DB.Exec(
		"INSERT INTO documents_fts (title, content, document_id) VALUES (?, ?, ?)",
		utils.Tokenize(d.Title), utils.Tokenize(d.Content), id,
	)
	if err != nil {
		return id, err
	}

	return id, nil
}

// GetDocumentByID 根据 ID 获取文档详情
func GetDocumentByID(id int64) (*Document, error) {
	d := &Document{}
	err := database.DB.QueryRow(
		"SELECT id, category_id, title, content, author_id, is_published, created_at, updated_at FROM documents WHERE id = ?",
		id,
	).Scan(&d.ID, &d.CategoryID, &d.Title, &d.Content, &d.AuthorID, &d.IsPublished, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return d, err
}

// ListDocuments 分页查询文档列表（含作者名）
func ListDocuments(page, size int, includeDraft bool) ([]DocumentListItem, int64, error) {
	where := "WHERE 1=1"
	if !includeDraft {
		where += " AND d.is_published = 1"
	}

	var total int64
	err := database.DB.QueryRow("SELECT COUNT(*) FROM documents d " + where).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	query := `SELECT d.id, d.category_id, d.title, d.author_id, u.username, d.is_published, d.created_at, d.updated_at
		FROM documents d LEFT JOIN users u ON d.author_id = u.id ` + where + ` ORDER BY d.updated_at DESC LIMIT ? OFFSET ?`

	rows, err := database.DB.Query(query, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var docs []DocumentListItem
	for rows.Next() {
		var d DocumentListItem
		if err := rows.Scan(&d.ID, &d.CategoryID, &d.Title, &d.AuthorID, &d.AuthorName, &d.IsPublished, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, err
		}
		docs = append(docs, d)
	}
	return docs, total, nil
}

// ListDocumentsByCategory 按分类查询已发布文档
func ListDocumentsByCategory(categoryID int64) ([]DocumentListItem, error) {
	rows, err := database.DB.Query(
		`SELECT d.id, d.category_id, d.title, d.author_id, u.username, d.is_published, d.created_at, d.updated_at
		FROM documents d LEFT JOIN users u ON d.author_id = u.id
		WHERE d.category_id = ? AND d.is_published = 1 ORDER BY d.updated_at DESC`,
		categoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []DocumentListItem
	for rows.Next() {
		var d DocumentListItem
		if err := rows.Scan(&d.ID, &d.CategoryID, &d.Title, &d.AuthorID, &d.AuthorName, &d.IsPublished, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, nil
}

// DocumentFilter 文档过滤条件
type DocumentFilter struct {
	CategoryID *int64
	Status     string // "published", "draft", or "" (all)
}

// ListDocumentsFiltered 带过滤条件的分页查询文档列表
func ListDocumentsFiltered(page, size int, includeDraft bool, filter DocumentFilter) ([]DocumentListItem, int64, error) {
	where := "WHERE 1=1"
	args := []interface{}{}

	if !includeDraft {
		where += " AND d.is_published = 1"
	} else if filter.Status == "published" {
		where += " AND d.is_published = 1"
	} else if filter.Status == "draft" {
		where += " AND d.is_published = 0"
	}

	if filter.CategoryID != nil {
		where += " AND d.category_id = ?"
		args = append(args, *filter.CategoryID)
	}

	var total int64
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := database.DB.QueryRow("SELECT COUNT(*) FROM documents d "+where, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	query := `SELECT d.id, d.category_id, d.title, d.author_id, u.username, d.is_published, d.created_at, d.updated_at
		FROM documents d LEFT JOIN users u ON d.author_id = u.id ` + where + ` ORDER BY d.updated_at DESC LIMIT ? OFFSET ?`
	args = append(args, size, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var docs []DocumentListItem
	for rows.Next() {
		var d DocumentListItem
		if err := rows.Scan(&d.ID, &d.CategoryID, &d.Title, &d.AuthorID, &d.AuthorName, &d.IsPublished, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, err
		}
		docs = append(docs, d)
	}
	return docs, total, nil
}

// UpdateDocument 更新文档内容
func UpdateDocument(d *Document) error {
	_, err := database.DB.Exec(
		"UPDATE documents SET title = ?, content = ?, category_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		d.Title, d.Content, d.CategoryID, d.ID,
	)
	if err != nil {
		return err
	}

	// 更新 FTS 索引（中文分词后索引）
	_, _ = database.DB.Exec("DELETE FROM documents_fts WHERE document_id = ?", d.ID)
	_, err = database.DB.Exec(
		"INSERT INTO documents_fts (title, content, document_id) VALUES (?, ?, ?)",
		utils.Tokenize(d.Title), utils.Tokenize(d.Content), d.ID,
	)
	return err
}

// PublishDocument 发布/取消发布文档
func PublishDocument(id int64, published bool) error {
	_, err := database.DB.Exec(
		"UPDATE documents SET is_published = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		published, id,
	)
	return err
}

// DeleteDocument 删除文档
func DeleteDocument(id int64) error {
	_, _ = database.DB.Exec("DELETE FROM documents_fts WHERE document_id = ?", id)
	_, err := database.DB.Exec("DELETE FROM documents WHERE id = ?", id)
	return err
}

// GetDocumentByTitle 根据标题查找文档
func GetDocumentByTitle(title string) (*Document, error) {
	d := &Document{}
	err := database.DB.QueryRow(
		"SELECT id, category_id, title, content, author_id, is_published, created_at, updated_at FROM documents WHERE title = ? LIMIT 1",
		title,
	).Scan(&d.ID, &d.CategoryID, &d.Title, &d.Content, &d.AuthorID, &d.IsPublished, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return d, err
}

// ListAllDocuments 列出所有文档（用于批量更新）
func ListAllDocuments() ([]Document, error) {
	rows, err := database.DB.Query("SELECT id, category_id, title, content, author_id, is_published, created_at, updated_at FROM documents")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.CategoryID, &d.Title, &d.Content, &d.AuthorID, &d.IsPublished, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

// UpdateDocumentContent 更新文档内容（含 FTS 索引同步）
func UpdateDocumentContent(id int64, content string) error {
	_, err := database.DB.Exec("UPDATE documents SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", content, id)
	if err != nil {
		return err
	}

	// 更新 FTS 索引
	var title string
	_ = database.DB.QueryRow("SELECT title FROM documents WHERE id = ?", id).Scan(&title)
	_, _ = database.DB.Exec("DELETE FROM documents_fts WHERE document_id = ?", id)
	_, err = database.DB.Exec(
		"INSERT INTO documents_fts (title, content, document_id) VALUES (?, ?, ?)",
		utils.Tokenize(title), utils.Tokenize(content), id,
	)
	return err
}
