// Package config 提供配置加载功能，从环境变量读取
package config

import (
	"os"
)

// Config 应用配置
type Config struct {
	Port              string // 监听端口
	DBPath            string // SQLite 数据库路径
	JWTSecret         string // JWT 签名密钥
	AdminInitPassword string // 初始管理员密码
	GinMode           string // Gin 运行模式
	UploadDir         string // 上传文件目录
}

// Load 从环境变量加载配置，提供合理默认值
func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		DBPath:            getEnv("DB_PATH", "./db/litewiki.db"),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-to-random-64-chars"),
		AdminInitPassword: getEnv("ADMIN_INIT_PASSWORD", "admin123456"),
		GinMode:           getEnv("GIN_MODE", "debug"),
		UploadDir:         getEnv("UPLOAD_DIR", "./uploads"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
