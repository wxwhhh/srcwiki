package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"litewiki/models"
	"litewiki/services"
	"litewiki/utils"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// AdminImport ZIP 导入处理器
type AdminImport struct {
	UploadDir string
}

// ImportProgress 导入进度事件
type ImportProgress struct {
	Stage        string `json:"stage"`
	Processed    int    `json:"processed,omitempty"`
	Total        int    `json:"total,omitempty"`
	CurrentFile  string `json:"current_file,omitempty"`
	Message      string `json:"message,omitempty"`
}

// ImportResult 导入结果
type ImportResult struct {
	CategoriesCreated int      `json:"categories_created"`
	CategoriesSkipped int      `json:"categories_skipped"`
	DocsCreated       int      `json:"docs_created"`
	DocsUpdated       int      `json:"docs_updated"`
	DocsSkipped       int      `json:"docs_skipped"`
	DuplicatesFound   int      `json:"duplicates_found"` // 重复文档数
	ImagesImported    int      `json:"images_imported"`
	Errors            []string `json:"errors"`
}

// Import 处理 ZIP 导入（SSE 流式响应）
func (a *AdminImport) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, 400, 40001, "请选择文件")
		return
	}

	// 大小限制 500MB
	if file.Size > 500*1024*1024 {
		utils.Error(c, 400, 40002, "文件大小不能超过 500MB")
		return
	}

	// 验证是 ZIP 文件
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".zip") {
		utils.Error(c, 400, 40003, "请上传 .zip 文件")
		return
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		utils.Error(c, 500, 50000, "打开文件失败")
		return
	}
	defer src.Close()

	// 读取到内存（ZIP 需要随机访问）
	zipData, err := io.ReadAll(src)
	if err != nil {
		utils.Error(c, 500, 50001, "读取文件失败")
		return
	}

	// 解析 ZIP
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		utils.Error(c, 400, 40004, "无效的 ZIP 文件")
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		utils.Error(c, 500, 50002, "不支持流式响应")
		return
	}

	// 发送 SSE 事件的辅助函数
	sendEvent := func(event ImportProgress) {
		data, _ := json.Marshal(event)
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		flusher.Flush()
	}

	// 第一阶段：扫描文件
	sendEvent(ImportProgress{Stage: "scanning", Message: "正在扫描 ZIP 文件..."})

	// 收集所有文件
	type zipEntry struct {
		Path string
		File *zip.File
	}

	var mdFiles []zipEntry
	var imgFiles []zipEntry
	ignoredPrefixes := []string{"__MACOSX/", ".DS_Store", "._"}

	for _, f := range zipReader.File {
		// 解码文件名（处理 GBK/GB2312 编码的 ZIP 文件名）
		fName := decodeZipFilename(f.Name, f.NonUTF8)

		// 跳过系统文件
		shouldSkip := false
		for _, prefix := range ignoredPrefixes {
			if strings.HasPrefix(fName, prefix) || strings.Contains(fName, "/"+prefix) {
				shouldSkip = true
				break
			}
		}
		if shouldSkip || f.FileInfo().IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(fName))
		switch ext {
		case ".md":
			// 检查大小限制 1MB
			if f.UncompressedSize64 > 1*1024*1024 {
				continue
			}
			mdFiles = append(mdFiles, zipEntry{Path: fName, File: f})
		case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg":
			imgFiles = append(imgFiles, zipEntry{Path: fName, File: f})
		}
	}

	sendEvent(ImportProgress{
		Stage:     "scanning",
		Processed: 0,
		Total:     len(mdFiles) + len(imgFiles),
		Message:   fmt.Sprintf("发现 %d 个 Markdown 文件，%d 个图片文件", len(mdFiles), len(imgFiles)),
	})

	// 第二阶段：导入
	result := &ImportResult{}
	createdCategories := make(map[string]int64) // 路径 -> 分类 ID

	// 先导入图片（建立路径映射）
	imageURLMap := make(map[string]string) // 原始路径 -> 上传后的 URL

	processed := 0
	for _, entry := range imgFiles {
		processed++
		sendEvent(ImportProgress{
			Stage:       "importing",
			Processed:   processed,
			Total:       len(mdFiles) + len(imgFiles),
			CurrentFile: entry.Path,
			Message:     "导入图片...",
		})

		// 读取图片内容
		rc, err := entry.File.Open()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("打开图片 %s 失败: %v", entry.Path, err))
			continue
		}

		ext := filepath.Ext(entry.Path)
		filename := time.Now().Format("20060102") + "/" + uuid.New().String() + ext
		dst := filepath.Join(a.UploadDir, filename)

		// 确保目录存在
		if err := utils.EnsureDir(filepath.Dir(dst)); err != nil {
			rc.Close()
			result.Errors = append(result.Errors, fmt.Sprintf("创建目录失败: %v", err))
			continue
		}

		// 写入文件
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("读取图片 %s 失败: %v", entry.Path, err))
			continue
		}

		if err := utils.WriteFile(dst, data); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("保存图片 %s 失败: %v", entry.Path, err))
			continue
		}

		imageURLMap[entry.Path] = "/uploads/" + filename
		// 也用文件名做索引（兼容相对路径引用）
		imageURLMap[filepath.Base(entry.Path)] = "/uploads/" + filename
		result.ImagesImported++
	}

	// 导入 Markdown 文件
	for _, entry := range mdFiles {
		processed++
		sendEvent(ImportProgress{
			Stage:       "importing",
			Processed:   processed,
			Total:       len(mdFiles) + len(imgFiles),
			CurrentFile: entry.Path,
			Message:     "导入文档...",
		})

		// 读取文件内容
		rc, err := entry.File.Open()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("打开文件 %s 失败: %v", entry.Path, err))
			continue
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("读取文件 %s 失败: %v", entry.Path, err))
			continue
		}

		// 编码检测和转换（GBK/GB2312 -> UTF-8）
		contentStr := detectAndConvertEncoding(content)

		// 清理语雀来源链接
		contentStr = utils.CleanYuqueLinks(contentStr)

		// 替换图片路径
		contentStr = replaceImagePaths(contentStr, imageURLMap, entry.Path)

		// 解析目录结构 -> 分类
		dir := filepath.Dir(entry.Path)
		// 去掉 ZIP 根目录前缀（如果有）
		parts := strings.Split(dir, "/")
		if len(parts) > 0 && parts[0] == "" {
			parts = parts[1:]
		}

		// 创建或获取分类
		var categoryID *int64
		if len(parts) > 0 && parts[0] != "." {
			catID, err := ensureCategoryPath(parts, createdCategories)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("创建分类 %s 失败: %v", dir, err))
			} else {
				categoryID = &catID
			}
		}

		// 文档标题 = 文件名（去掉 .md 后缀）
		title := strings.TrimSuffix(filepath.Base(entry.Path), ".md")

		// 查找是否已存在同标题文档
		existingDoc, _ := findDocumentByTitle(title)
		if existingDoc != nil {
			// 更新
			existingDoc.Content = contentStr
			if categoryID != nil {
				existingDoc.CategoryID = categoryID
			}
			if err := models.UpdateDocument(existingDoc); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("更新文档 %s 失败: %v", title, err))
			} else {
				result.DocsUpdated++
			}
		} else {
			// 创建新文档（默认发布）
			doc := &models.Document{
				Title:       title,
				Content:     contentStr,
				CategoryID:  categoryID,
				AuthorID:    1, // 默认管理员
				IsPublished: true,
			}
			if _, err := models.CreateDocument(doc); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("创建文档 %s 失败: %v", title, err))
			} else {
				result.DocsCreated++
			}
		}
	}

	// 发送最终结果
	resultJSON, _ := json.Marshal(result)
	sendEvent(ImportProgress{
		Stage:   "result",
		Message: string(resultJSON),
	})

	sendEvent(ImportProgress{Stage: "done"})

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "import_zip",
		TargetType: "import",
		Detail:     fmt.Sprintf("categories:%d, docs:%d, images:%d", result.CategoriesCreated+result.CategoriesSkipped, result.DocsCreated+result.DocsUpdated, result.ImagesImported),
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
}

// ensureCategoryPath 确保分类路径存在，返回最深层分类 ID
func ensureCategoryPath(parts []string, cache map[string]int64) (int64, error) {
	if len(parts) == 0 {
		return 0, fmt.Errorf("空路径")
	}

	// 构建完整路径键
	fullPath := strings.Join(parts, "/")
	if id, ok := cache[fullPath]; ok {
		return id, nil
	}

	var parentID *int64
	for i, part := range parts {
		currentPath := strings.Join(parts[:i+1], "/")
		if id, ok := cache[currentPath]; ok {
			parentID = &id
			continue
		}

		// 查找或创建分类
		cat, err := services.FindCategoryByNameAndParent(part, parentID)
		if err != nil || cat == nil {
			// 创建新分类
			newCat, err := services.CreateCategory(&services.CreateCategoryRequest{
				Name:     part,
				ParentID: parentID,
			})
			if err != nil {
				return 0, fmt.Errorf("创建分类 %s 失败: %w", part, err)
			}
			cat = newCat
		}

		cache[currentPath] = cat.ID
		parentID = &cat.ID
	}

	if parentID == nil {
		return 0, fmt.Errorf("无法创建分类")
	}
	return *parentID, nil
}

// findDocumentByTitle 根据标题查找文档
func findDocumentByTitle(title string) (*models.Document, error) {
	return models.GetDocumentByTitle(title)
}

// detectAndConvertEncoding 检测并转换编码为 UTF-8
func detectAndConvertEncoding(data []byte) string {
	// 先检查是否已经是有效的 UTF-8
	if utf8.Valid(data) {
		return string(data)
	}
	// 如果不是 UTF-8，尝试 GBK 转换
	if isLikelyGBK(data) {
		reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder())
		converted, err := io.ReadAll(reader)
		if err == nil && utf8.Valid(converted) {
			return string(converted)
		}
	}
	// GB18030 是 GBK 的超集，最后尝试
	reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GB18030.NewDecoder())
	converted, err := io.ReadAll(reader)
	if err == nil && utf8.Valid(converted) {
		return string(converted)
	}
	// 都失败了，返回原始数据
	return string(data)
}

// isLikelyGBK 简单判断是否可能是 GBK 编码
func isLikelyGBK(data []byte) bool {
	for i := 0; i < len(data)-1; i++ {
		if data[i] >= 0x81 && data[i] <= 0xFE {
			if data[i+1] >= 0x40 && data[i+1] <= 0xFE && data[i+1] != 0x7F {
				return true
			}
		}
	}
	return false
}

// decodeZipFilename 解码 ZIP 文件名编码
// Windows 创建的 ZIP 文件名通常是 GBK 编码，Go 的 archive/zip 不会自动转换
func decodeZipFilename(name string, nonUTF8 bool) string {
	// 如果标记为 UTF-8 且确实是有效 UTF-8，直接返回
	if !nonUTF8 && utf8.ValidString(name) {
		return name
	}
	// 尝试 GBK 解码
	reader := transform.NewReader(bytes.NewReader([]byte(name)), simplifiedchinese.GBK.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err == nil && utf8.Valid(decoded) {
		return string(decoded)
	}
	// GB18030 是 GBK 的超集，再尝试一次
	reader = transform.NewReader(bytes.NewReader([]byte(name)), simplifiedchinese.GB18030.NewDecoder())
	decoded, err = io.ReadAll(reader)
	if err == nil && utf8.Valid(decoded) {
		return string(decoded)
	}
	// 都失败了，返回原始值
	return name
}

// replaceImagePaths 替换 Markdown 中的图片路径
func replaceImagePaths(content string, urlMap map[string]string, mdPath string) string {
	result := content
	mdDir := filepath.Dir(mdPath)

	// 遍历 urlMap，找到所有可能的引用方式
	for origPath, newURL := range urlMap {
		// origPath 是 ZIP 内的完整路径，如 "学校/公司名/img/xxx/file.webp"
		// mdPath 是 md 文件的 ZIP 内路径，如 "学校/公司名/file.md"
		// mdDir 是 md 文件所在目录，如 "学校/公司名"

		// 1. 如果 origPath 以 mdDir 开头，构建相对路径
		if strings.HasPrefix(origPath, mdDir+"/") {
			relPath := strings.TrimPrefix(origPath, mdDir+"/")
			// 替换 ./relPath 和 relPath
			result = strings.ReplaceAll(result, "(./"+relPath+")", "("+newURL+")")
			result = strings.ReplaceAll(result, "("+relPath+")", "("+newURL+")")
		}

		// 2. 用文件名匹配（兼容各种引用方式）
		baseName := filepath.Base(origPath)
		result = strings.ReplaceAll(result, "(./img/"+baseName+")", "("+newURL+")")
		result = strings.ReplaceAll(result, "(img/"+baseName+")", "("+newURL+")")
		result = strings.ReplaceAll(result, "("+baseName+")", "("+newURL+")")
	}

	return result
}
