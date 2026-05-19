// Package utils 提供文件操作工具函数
package utils

import (
	"os"
	"path/filepath"
)

// EnsureDir 确保目录存在
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// WriteFile 写入文件（自动创建目录）
func WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
