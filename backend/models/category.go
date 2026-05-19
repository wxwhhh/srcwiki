package models

import (
	"database/sql"
	"errors"
	"litewiki/database"
	"time"
)

// Category 分类模型
type Category struct {
	ID        int64     `json:"id"`
	ParentID  *int64    `json:"parent_id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	DocCount  int       `json:"doc_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TreeDoc 树节点文档（轻量）
type TreeDoc struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	IsPublished bool   `json:"is_published"`
}

// CategoryTreeNode 分类树节点（含文档数 + 文档列表）
type CategoryTreeNode struct {
	Category
	DocCount int               `json:"doc_count"`
	Docs     []TreeDoc         `json:"docs,omitempty"`
	Children []CategoryTreeNode `json:"children,omitempty"`
}

// CreateCategory 创建分类
func CreateCategory(c *Category) (int64, error) {
	result, err := database.DB.Exec(
		"INSERT INTO categories (parent_id, name, sort_order) VALUES (?, ?, ?)",
		c.ParentID, c.Name, c.SortOrder,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetCategoryByID 根据 ID 查找分类
func GetCategoryByID(id int64) (*Category, error) {
	c := &Category{}
	err := database.DB.QueryRow(
		"SELECT id, parent_id, name, sort_order, created_at, updated_at FROM categories WHERE id = ?",
		id,
	).Scan(&c.ID, &c.ParentID, &c.Name, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

// ListCategories 查询所有分类（平铺，含文档数）
func ListCategories() ([]Category, error) {
	rows, err := database.DB.Query(
		"SELECT id, parent_id, name, sort_order, created_at, updated_at FROM categories ORDER BY sort_order ASC, id ASC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Name, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	// 查询每个分类的文档数
	if len(categories) > 0 {
		docCounts := make(map[int64]int)
		countRows, err := database.DB.Query(
			"SELECT category_id, COUNT(*) FROM documents GROUP BY category_id",
		)
		if err == nil {
			defer countRows.Close()
			for countRows.Next() {
				var catID sql.NullInt64
				var count int
				if err := countRows.Scan(&catID, &count); err == nil && catID.Valid {
					docCounts[catID.Int64] = count
				}
			}
		}
		for i := range categories {
			categories[i].DocCount = docCounts[categories[i].ID]
		}
	}

	return categories, nil
}

// GetCategoryTree 获取分类树（含文档数）
func GetCategoryTree() ([]CategoryTreeNode, error) {
	cats, err := ListCategories()
	if err != nil {
		return nil, err
	}

	// 查询每个分类的文档数（仅已发布）
	docCounts := make(map[int64]int)
	rows, err := database.DB.Query(
		"SELECT category_id, COUNT(*) FROM documents WHERE is_published = 1 GROUP BY category_id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var catID sql.NullInt64
		var count int
		if err := rows.Scan(&catID, &count); err != nil {
			return nil, err
		}
		if catID.Valid {
			docCounts[catID.Int64] = count
		}
	}

	// 构建树
	nodes := make([]CategoryTreeNode, len(cats))
	for i, c := range cats {
		nodes[i] = CategoryTreeNode{Category: c, DocCount: docCounts[c.ID]}
	}

	result := buildTree(nodes, nil)

	// 查询每个分类的文档列表（仅叶子节点需要，但全量填充更简单）
	docRows, err := database.DB.Query(
		"SELECT id, title, is_published, category_id FROM documents WHERE is_published = 1 ORDER BY title ASC",
	)
	if err != nil {
		return result, nil // 文档查询失败不影响树结构
	}
	defer docRows.Close()

	catDocs := make(map[int64][]TreeDoc)
	for docRows.Next() {
		var doc TreeDoc
		var catID sql.NullInt64
		var isPub int
		if err := docRows.Scan(&doc.ID, &doc.Title, &isPub, &catID); err != nil {
			continue
		}
		doc.IsPublished = isPub == 1
		if catID.Valid {
			catDocs[catID.Int64] = append(catDocs[catID.Int64], doc)
		}
	}

	// 将文档列表填充到各节点
	fillDocs(result, catDocs)

	return result, nil
}

// buildTree 递归构建分类树
func buildTree(nodes []CategoryTreeNode, parentID *int64) []CategoryTreeNode {
	var result []CategoryTreeNode
	for _, node := range nodes {
		if (parentID == nil && node.ParentID == nil) ||
			(parentID != nil && node.ParentID != nil && *node.ParentID == *parentID) {
			node.Children = buildTree(nodes, &node.ID)
			result = append(result, node)
		}
	}
	return result
}

// fillDocs 递归将文档列表填充到各节点
func fillDocs(nodes []CategoryTreeNode, catDocs map[int64][]TreeDoc) {
	for i := range nodes {
		if docs, ok := catDocs[nodes[i].ID]; ok {
			nodes[i].Docs = docs
		}
		if len(nodes[i].Children) > 0 {
			fillDocs(nodes[i].Children, catDocs)
		}
	}
}

// buildCategoryTree 从平铺分类构建树
func buildCategoryTree(cats []Category, parentID *int64) []Category {
	var result []Category
	for _, c := range cats {
		if (parentID == nil && c.ParentID == nil) ||
			(parentID != nil && c.ParentID != nil && *c.ParentID == *parentID) {
			result = append(result, c)
		}
	}
	return result
}

// UpdateCategory 更新分类
func UpdateCategory(c *Category) error {
	_, err := database.DB.Exec(
		"UPDATE categories SET name = ?, parent_id = ?, sort_order = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		c.Name, c.ParentID, c.SortOrder, c.ID,
	)
	return err
}

// DeleteCategory 删除分类（级联删除子分类）
func DeleteCategory(id int64) error {
	_, err := database.DB.Exec("DELETE FROM categories WHERE id = ?", id)
	return err
}

// GetCategoryByNameAndParent 根据名称和父分类查找分类
func GetCategoryByNameAndParent(name string, parentID *int64) (*Category, error) {
	c := &Category{}
	var err error
	if parentID == nil {
		err = database.DB.QueryRow(
			"SELECT id, parent_id, name, sort_order, created_at, updated_at FROM categories WHERE name = ? AND parent_id IS NULL",
			name,
		).Scan(&c.ID, &c.ParentID, &c.Name, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt)
	} else {
		err = database.DB.QueryRow(
			"SELECT id, parent_id, name, sort_order, created_at, updated_at FROM categories WHERE name = ? AND parent_id = ?",
			name, *parentID,
		).Scan(&c.ID, &c.ParentID, &c.Name, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt)
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

// GetAllChildCategoryIDs 递归获取所有子分类 ID（使用 SQL 递归 CTE）
func GetAllChildCategoryIDs(parentID int64) ([]int64, error) {
	var ids []int64
	rows, err := database.DB.Query(`
		WITH RECURSIVE children AS (
			SELECT id FROM categories WHERE parent_id = ?
			UNION ALL
			SELECT c.id FROM categories c INNER JOIN children ch ON c.parent_id = ch.id
		)
		SELECT id FROM children
	`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// CascadeDeleteCategories 级联批量删除分类及其下所有文档（事务）
func CascadeDeleteCategories(ids []int64) (deletedCats int, deletedDocs int, err error) {
	if len(ids) == 0 {
		return 0, 0, errors.New("请选择要删除的分类")
	}

	// 使用 SQL 递归 CTE 收集所有子分类 ID（避免单连接死锁）
	allCatIDs := make(map[int64]bool)
	for _, id := range ids {
		allCatIDs[id] = true
		childIDs, err := GetAllChildCategoryIDs(id)
		if err != nil {
			return 0, 0, err
		}
		for _, cid := range childIDs {
			allCatIDs[cid] = true
		}
	}

	catIDList := make([]int64, 0, len(allCatIDs))
	for id := range allCatIDs {
		catIDList = append(catIDList, id)
	}
	if len(catIDList) == 0 {
		return 0, 0, nil
	}

	placeholders := ""
	args := make([]interface{}, len(catIDList))
	for i, id := range catIDList {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	// 1. 删除 FTS 索引
	_, err = tx.Exec(
		"DELETE FROM documents_fts WHERE document_id IN (SELECT id FROM documents WHERE category_id IN ("+placeholders+"))",
		args...,
	)
	if err != nil {
		return 0, 0, err
	}

	// 2. 删除关联文档
	result, err := tx.Exec(
		"DELETE FROM documents WHERE category_id IN ("+placeholders+")",
		args...,
	)
	if err != nil {
		return 0, 0, err
	}
	docsAffected, _ := result.RowsAffected()

	// 3. 删除分类
	result, err = tx.Exec(
		"DELETE FROM categories WHERE id IN ("+placeholders+")",
		args...,
	)
	if err != nil {
		return 0, 0, err
	}
	catsAffected, _ := result.RowsAffected()

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}

	return int(catsAffected), int(docsAffected), nil
}

// UpdateCategorySort 批量更新分类排序
func UpdateCategorySort(ids []int64) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range ids {
		_, err := tx.Exec("UPDATE categories SET sort_order = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", i, id)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
