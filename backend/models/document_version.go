package models

import (
	"litewiki/database"
	"time"
)

// DocumentVersion 文档版本模型
type DocumentVersion struct {
	ID         int64     `json:"id"`
	DocumentID int64     `json:"document_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	EditorID   int64     `json:"editor_id"`
	EditorName string    `json:"editor_name"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateDocumentVersion 创建文档版本记录
func CreateDocumentVersion(v *DocumentVersion) (int64, error) {
	result, err := database.DB.Exec(
		"INSERT INTO document_versions (document_id, title, content, editor_id) VALUES (?, ?, ?, ?)",
		v.DocumentID, v.Title, v.Content, v.EditorID,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// ListDocumentVersions 获取文档版本历史
func ListDocumentVersions(documentID int64) ([]DocumentVersion, error) {
	rows, err := database.DB.Query(
		`SELECT v.id, v.document_id, v.title, v.content, v.editor_id, u.username, v.created_at
		FROM document_versions v LEFT JOIN users u ON v.editor_id = u.id
		WHERE v.document_id = ? ORDER BY v.created_at DESC`,
		documentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []DocumentVersion
	for rows.Next() {
		var v DocumentVersion
		if err := rows.Scan(&v.ID, &v.DocumentID, &v.Title, &v.Content, &v.EditorID, &v.EditorName, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}

// GetDocumentVersionByID 根据 ID 获取版本
func GetDocumentVersionByID(id int64) (*DocumentVersion, error) {
	v := &DocumentVersion{}
	err := database.DB.QueryRow(
		`SELECT v.id, v.document_id, v.title, v.content, v.editor_id, u.username, v.created_at
		FROM document_versions v LEFT JOIN users u ON v.editor_id = u.id WHERE v.id = ?`,
		id,
	).Scan(&v.ID, &v.DocumentID, &v.Title, &v.Content, &v.EditorID, &v.EditorName, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return v, nil
}
