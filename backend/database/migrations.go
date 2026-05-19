package database

import "github.com/rs/zerolog/log"

const createTablesSQL = `
-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    username    TEXT NOT NULL UNIQUE,
    password    TEXT NOT NULL,
    role        TEXT NOT NULL DEFAULT 'reader' CHECK(role IN ('admin','editor','reader')),
    status      TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active','disabled')),
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 邀请码表
CREATE TABLE IF NOT EXISTS invite_codes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    code        TEXT NOT NULL UNIQUE,
    role        TEXT NOT NULL DEFAULT 'reader' CHECK(role IN ('editor','reader')),
    max_uses    INTEGER DEFAULT 1,
    use_count   INTEGER DEFAULT 0,
    expires_at  DATETIME,
    created_by  INTEGER REFERENCES users(id),
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 分类表
CREATE TABLE IF NOT EXISTS categories (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id   INTEGER REFERENCES categories(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    sort_order  INTEGER DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 文档表
CREATE TABLE IF NOT EXISTS documents (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id  INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    title        TEXT NOT NULL,
    content      TEXT NOT NULL DEFAULT '',
    author_id    INTEGER REFERENCES users(id),
    is_published BOOLEAN DEFAULT FALSE,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 文档版本表
CREATE TABLE IF NOT EXISTS document_versions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    document_id INTEGER NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    content     TEXT NOT NULL,
    editor_id   INTEGER REFERENCES users(id),
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 导入任务表
CREATE TABLE IF NOT EXISTS import_tasks (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    type          TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    source        TEXT,
    progress      INTEGER DEFAULT 0,
    total_docs    INTEGER DEFAULT 0,
    imported_docs INTEGER DEFAULT 0,
    updated_docs  INTEGER DEFAULT 0,
    skipped_docs  INTEGER DEFAULT 0,
    error_count   INTEGER DEFAULT 0,
    errors        TEXT,
    result        TEXT,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    started_at    DATETIME,
    finished_at   DATETIME
);

-- 致谢表
CREATE TABLE IF NOT EXISTS credits (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    url         TEXT NOT NULL,
    description TEXT,
    icon_url    TEXT,
    license     TEXT,
    stars       TEXT,
    sort_order  INTEGER DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 系统选项表
CREATE TABLE IF NOT EXISTS options (
    key         TEXT PRIMARY KEY,
    value       TEXT NOT NULL DEFAULT '',
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_log (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER REFERENCES users(id),
    username    TEXT,
    action      TEXT NOT NULL,
    target_type TEXT,
    target_id   INTEGER,
    detail      TEXT,
    ip          TEXT,
    user_agent  TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 全文搜索虚拟表
CREATE VIRTUAL TABLE IF NOT EXISTS documents_fts USING fts5(
    title, content, document_id UNINDEXED
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_categories_parent ON categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_documents_category ON documents(category_id);
CREATE INDEX IF NOT EXISTS idx_documents_published ON documents(is_published);
CREATE INDEX IF NOT EXISTS idx_audit_log_user ON audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_created ON audit_log(created_at);
`

// Migrate 执行数据库迁移
func Migrate() error {
	_, err := DB.Exec(createTablesSQL)
	if err != nil {
		return err
	}

	// 初始化默认系统选项（INSERT OR IGNORE 避免覆盖已有值）
	_, err = DB.Exec("INSERT OR IGNORE INTO options (key, value) VALUES ('register_mode', 'invite')")
	if err != nil {
		return err
	}

	log.Info().Msg("数据库迁移完成")
	return nil
}

// RebuildFTSIndex 重建全文搜索索引（使用中文分词）
func RebuildFTSIndex(tokenizeFn func(string) string) error {
	log.Info().Msg("开始重建 FTS 索引...")

	// 清空 FTS 索引
	_, err := DB.Exec("DELETE FROM documents_fts")
	if err != nil {
		return err
	}

	// 查询所有文档（先收集到内存，避免连接死锁）
	rows, err := DB.Query("SELECT id, title, content FROM documents")
	if err != nil {
		return err
	}

	type docRow struct {
		id      int64
		title   string
		content string
	}
	var docs []docRow
	for rows.Next() {
		var d docRow
		if err := rows.Scan(&d.id, &d.title, &d.content); err != nil {
			rows.Close()
			return err
		}
		docs = append(docs, d)
	}
	rows.Close()

	// 批量插入 FTS 索引
	count := 0
	for _, d := range docs {
		_, err := DB.Exec(
			"INSERT INTO documents_fts (title, content, document_id) VALUES (?, ?, ?)",
			tokenizeFn(d.title), tokenizeFn(d.content), d.id,
		)
		if err != nil {
			return err
		}
		count++
	}

	log.Info().Int("count", count).Msg("FTS 索引重建完成")
	return nil
}

// CleanExistingYuqueLinks 清理已导入文档中的语雀来源链接，并重建受影响文档的 FTS 索引
func CleanExistingYuqueLinks(cleanFn func(string) string, tokenizeFn func(string) string) error {
	log.Info().Msg("开始清理文档中的语雀链接...")

	rows, err := DB.Query("SELECT id, title, content FROM documents")
	if err != nil {
		return err
	}

	type docRow struct {
		id      int64
		title   string
		content string
	}
	var docs []docRow
	for rows.Next() {
		var d docRow
		if err := rows.Scan(&d.id, &d.title, &d.content); err != nil {
			rows.Close()
			return err
		}
		docs = append(docs, d)
	}
	rows.Close()

	cleaned := 0
	for _, d := range docs {
		newContent := cleanFn(d.content)
		if newContent != d.content {
			_, err := DB.Exec("UPDATE documents SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", newContent, d.id)
			if err != nil {
				log.Warn().Err(err).Int64("id", d.id).Msg("清理文档失败")
				continue
			}
			// 更新 FTS 索引
			_, _ = DB.Exec("DELETE FROM documents_fts WHERE document_id = ?", d.id)
			_, _ = DB.Exec(
				"INSERT INTO documents_fts (title, content, document_id) VALUES (?, ?, ?)",
				tokenizeFn(d.title), tokenizeFn(newContent), d.id,
			)
			cleaned++
		}
	}

	if cleaned > 0 {
		log.Info().Int("count", cleaned).Msg("语雀链接清理完成")
	} else {
		log.Info().Msg("无需清理语雀链接")
	}
	return nil
}
