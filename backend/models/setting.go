package models

import (
	"database/sql"
	"litewiki/database"
)

// GetOption 获取系统选项值
func GetOption(key string) (string, error) {
	var value string
	err := database.DB.QueryRow("SELECT value FROM options WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetOption 设置系统选项值
func SetOption(key, value string) error {
	_, err := database.DB.Exec(
		"INSERT INTO options (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP",
		key, value,
	)
	return err
}

// GetRegisterMode 获取注册模式（open / invite），默认 invite
func GetRegisterMode() string {
	val, err := GetOption("register_mode")
	if err != nil || val == "" {
		return "invite"
	}
	if val != "open" && val != "invite" {
		return "invite"
	}
	return val
}
