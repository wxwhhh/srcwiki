package database

import (
	"litewiki/utils"

	"github.com/rs/zerolog/log"
)

// Seed 创建初始管理员账号（如果不存在）
func Seed(adminPassword string) error {
	// 检查是否已有管理员
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Info().Msg("管理员账号已存在，跳过 seed")
		return nil
	}

	// 创建默认管理员
	hash, err := utils.HashPassword(adminPassword)
	if err != nil {
		return err
	}

	_, err = DB.Exec(
		"INSERT INTO users (username, password, role, status) VALUES (?, ?, 'admin', 'active')",
		"admin", hash,
	)
	if err != nil {
		return err
	}

	log.Info().Msg("初始管理员账号创建成功 (admin)")
	return nil
}
