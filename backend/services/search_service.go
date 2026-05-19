package services

import (
	"database/sql"
	"litewiki/database"
	"litewiki/utils"
	"strings"
)

// SearchResult 搜索结果项
type SearchResult struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content_snippet"` // 摘要片段
	Category *int64 `json:"category_id"`
}

// Search 全文搜索（FTS5）
func Search(query string, page, size int) ([]SearchResult, int64, error) {
	if len(query) == 0 {
		return nil, 0, nil
	}
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	// 中文分词后构建 FTS5 查询
	tokenized := utils.TokenizeForSearch(query)
	words := strings.Fields(tokenized)
	var ftsParts []string
	for _, w := range words {
		w = strings.TrimSpace(w)
		if len(w) > 0 {
			ftsParts = append(ftsParts, w+"*")
		}
	}
	ftsQuery := strings.Join(ftsParts, " AND ")
	if ftsQuery == "" {
		ftsQuery = query + "*"
	}

	var total int64
	err := database.DB.QueryRow(
		`SELECT COUNT(*) FROM documents_fts
		WHERE documents_fts MATCH ? AND document_id IN (SELECT id FROM documents WHERE is_published = 1)`,
		ftsQuery,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	rows, err := database.DB.Query(
		`SELECT d.id, d.title, d.content, d.category_id
		FROM documents_fts fts
		JOIN documents d ON fts.document_id = d.id
		WHERE documents_fts MATCH ? AND d.is_published = 1
		ORDER BY rank
		LIMIT ? OFFSET ?`,
		ftsQuery, size, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var catID sql.NullInt64
		var fullContent string
		if err := rows.Scan(&r.ID, &r.Title, &fullContent, &catID); err != nil {
			return nil, 0, err
		}
		if catID.Valid {
			r.Category = &catID.Int64
		}
		// 从原始内容中截取摘要片段（而非使用 FTS snippet，避免显示分词结果）
		r.Content = buildSnippet(fullContent, query, 200)
		results = append(results, r)
	}
	return results, total, nil
}

// buildSnippet 从原始内容中提取包含关键词的摘要片段
func buildSnippet(content, query string, maxLen int) string {
	clean := stripMarkdown(content)

	idx := strings.Index(clean, query)
	if idx < 0 {
		runes := []rune(clean)
		if len(runes) > maxLen {
			return string(runes[:maxLen]) + "..."
		}
		return clean
	}

	start := idx - 80
	if start < 0 {
		start = 0
	}
	end := start + maxLen
	runes := []rune(clean)
	if end > len(runes) {
		end = len(runes)
	}
	if start > len(runes) {
		start = len(runes)
	}

	snippet := string(runes[start:end])
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(runes) {
		snippet = snippet + "..."
	}
	return snippet
}

// stripMarkdown 简单去除 markdown 和 HTML 标记，提取纯文本
func stripMarkdown(s string) string {
	var result strings.Builder
	inTag := false
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '<' {
			inTag = true
			continue
		}
		if inTag {
			if r == '>' {
				inTag = false
			}
			continue
		}
		// 跳过 markdown 标记符号
		if r == '#' || r == '*' || r == '_' || r == '`' || r == '[' || r == ']' || r == '!' || r == '>' {
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}
