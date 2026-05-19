// Package database 提供数据库连接和初始化
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
	"github.com/rs/zerolog/log"
)

var DB *sql.DB

// Init 初始化数据库连接，开启 WAL 模式
func Init(dbPath string) error {
	// 确保数据库目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建数据库目录失败: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON")
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	// 测试连接
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	// 设置连接池参数（SQLite 单写多读）
	DB.SetMaxOpenConns(1)

	log.Info().Str("path", dbPath).Msg("数据库连接成功")
	return nil
}

// Close 关闭数据库连接
func Close() {
	if DB != nil {
		DB.Close()
	}
}
