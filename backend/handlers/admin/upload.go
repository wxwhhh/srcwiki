package handlers

import (
	"litewiki/models"
	"litewiki/utils"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminUpload 文件上传处理器
type AdminUpload struct {
	UploadDir string
}

// Upload 上传图片/文件
func (a *AdminUpload) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, 400, 40001, "请选择文件")
		return
	}

	// 大小限制 5MB
	if file.Size > 5*1024*1024 {
		utils.Error(c, 400, 40002, "文件大小不能超过 5MB")
		return
	}

	// 白名单后缀
	ext := filepath.Ext(file.Filename)
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".webp": true, ".svg": true, ".pdf": true,
		".md": true, ".txt": true,
	}
	if !allowed[ext] {
		utils.Error(c, 400, 40003, "不支持的文件类型")
		return
	}

	// 生成随机文件名
	filename := time.Now().Format("20060102") + "/" + uuid.New().String() + ext
	dst := filepath.Join(a.UploadDir, filename)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		utils.Error(c, 500, 50000, "创建目录失败")
		return
	}

	if err := c.SaveUploadedFile(file, dst); err != nil {
		utils.Error(c, 500, 50001, "保存文件失败")
		return
	}

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "upload_file",
		TargetType: "file",
		Detail:     filename,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, gin.H{
		"url":      "/uploads/" + filename,
		"filename": filename,
	})
}
