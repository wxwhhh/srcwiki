package handlers

import (
	"litewiki/utils"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// UploadsHandler 处理上传文件的访问（需要鉴权）
type UploadsHandler struct {
	UploadDir string
}

// ServeFile 提供上传文件的访问（需要登录）
func (h *UploadsHandler) ServeFile(c *gin.Context) {
	// 从 URL 获取文件路径
	filePath := c.Param("filepath")
	if filePath == "" {
		utils.Error(c, http.StatusBadRequest, 40001, "文件路径不能为空")
		return
	}

	// 清理路径，防止路径遍历攻击
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		utils.Error(c, http.StatusBadRequest, 40002, "非法文件路径")
		return
	}

	// 构建完整文件路径
	fullPath := filepath.Join(h.UploadDir, cleanPath)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		utils.Error(c, http.StatusNotFound, 40400, "文件不存在")
		return
	}

	// 提供文件
	c.File(fullPath)
}
