package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"litewiki/models"
	"litewiki/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GitHubImportRequest GitHub 导入请求
type GitHubImportRequest struct {
	URL      string `json:"url" binding:"required"`
	Branch   string `json:"branch"`
	SkipRoot bool   `json:"skip_root"`
}

// GitHubImport GitHub 导入处理器
type GitHubImport struct{}

// githubURLToRaw 将 GitHub URL 转换为 raw URL
func githubURLToRaw(url string) string {
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimPrefix(url, "https://github.com/")
	return "https://raw.githubusercontent.com/" + url
}

// imageRef 图片引用信息
type imageRef struct {
	AbsPath  string // 仓库中的绝对路径
	Original string // markdown 中的原始引用
}

// extractImageRefs 从 markdown 内容中提取图片引用，dir 是文档在仓库中的目录
func extractImageRefs(content string, dir string) []imageRef {
	var refs []imageRef
	seen := make(map[string]bool)

	// 匹配 ![alt](path) 和 ![alt](path "title")
	re := regexp.MustCompile(`!\[([^\]]*)\]\(([^)\s]+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			origPath := m[2]
			// 跳过 HTTP URL 和锚点
			if strings.HasPrefix(origPath, "http://") || strings.HasPrefix(origPath, "https://") || strings.HasPrefix(origPath, "#") {
				continue
			}

			// 解析为仓库中的绝对路径
			absPath := resolveRepoPath(origPath, dir)
			key := absPath + "|" + origPath
			if !seen[key] {
				seen[key] = true
				refs = append(refs, imageRef{AbsPath: absPath, Original: origPath})
			}
		}
	}

	// 也匹配 HTML img 标签
	imgRe := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`)
	imgMatches := imgRe.FindAllStringSubmatch(content, -1)
	for _, m := range imgMatches {
		if len(m) >= 2 {
			origPath := m[1]
			if strings.HasPrefix(origPath, "http://") || strings.HasPrefix(origPath, "https://") || strings.HasPrefix(origPath, "#") {
				continue
			}
			absPath := resolveRepoPath(origPath, dir)
			key := absPath + "|" + origPath
			if !seen[key] {
				seen[key] = true
				refs = append(refs, imageRef{AbsPath: absPath, Original: origPath})
			}
		}
	}

	return refs
}

// resolveRepoPath 将相对路径解析为仓库根目录下的绝对路径
// 例如: docDir="泛微OA", relPath="./img/xxx.png" → "泛微OA/img/xxx.png"
//        docDir="泛微OA", relPath="../img/xxx.png" → "img/xxx.png"
func resolveRepoPath(relPath, docDir string) string {
	// 清理路径
	if docDir == "" || docDir == "." {
		return filepath.Clean(relPath)
	}
	combined := filepath.Join(docDir, relPath)
	return filepath.Clean(combined)
}

// sanitizeFilename 清理文件名
func sanitizeFilename(name string) string {
	name = strings.TrimSuffix(name, filepath.Ext(name))
	var result []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || (r >= 0x4e00 && r <= 0x9fff) {
			result = append(result, r)
		} else {
			result = append(result, '_')
		}
	}
	return string(result)
}

// Import 处理 GitHub 仓库导入（SSE 流式响应，使用 git sparse-checkout）
func (g *GitHubImport) Import(c *gin.Context) {
	var req GitHubImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40001, "请求参数错误")
		return
	}

	if req.Branch == "" {
		out, err := exec.Command("git", "ls-remote", "--symref", req.URL, "HEAD").CombinedOutput()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				if strings.HasPrefix(line, "ref: refs/heads/") {
					parts := strings.SplitN(line, "\t", 2)
					branch := strings.TrimPrefix(parts[0], "ref: refs/heads/")
					if branch != "" {
						req.Branch = branch
					}
				}
			}
		}
		if req.Branch == "" {
			req.Branch = "main"
		}
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		utils.Error(c, 500, 50002, "不支持流式响应")
		return
	}

	sendEvent := func(event ImportProgress) {
		data, _ := json.Marshal(event)
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		flusher.Flush()
	}

	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("github_import_%s", uuid.New().String()[:8]))
	defer os.RemoveAll(tmpDir)

	// 阶段 1：git clone
	sendEvent(ImportProgress{Stage: "cloning", Message: "正在克隆仓库（仅下载 .md 文件）..."})

	cloneArgs := []string{
		"clone",
		"--filter=blob:none",
		"--sparse",
		"--depth", "1",
		"--single-branch",
		"--branch", req.Branch,
		req.URL, tmpDir,
	}

	cloneCmd := exec.Command("git", cloneArgs...)
	cloneCmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cloneCmd.CombinedOutput()
	if err != nil {
		errMsg := string(output)
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "Repository not found") {
			sendEvent(ImportProgress{Stage: "error", Message: "仓库不存在，请检查 URL"})
		} else if strings.Contains(errMsg, "Could not find remote branch") {
			sendEvent(ImportProgress{Stage: "error", Message: fmt.Sprintf("分支 '%s' 不存在", req.Branch)})
		} else {
			sendEvent(ImportProgress{Stage: "error", Message: fmt.Sprintf("克隆失败: %s", truncate(errMsg, 200))})
		}
		return
	}

	// 阶段 2：sparse-checkout 只检出 .md
	sendEvent(ImportProgress{Stage: "scanning", Message: "正在检出 .md 文件..."})

	initCmd := exec.Command("git", "sparse-checkout", "init", "--no-cone")
	initCmd.Dir = tmpDir
	if output, err := initCmd.CombinedOutput(); err != nil {
		sendEvent(ImportProgress{Stage: "error", Message: fmt.Sprintf("初始化稀疏检出失败: %s", string(output))})
		return
	}

	checkoutCmd := exec.Command("git", "sparse-checkout", "set", "*.md")
	checkoutCmd.Dir = tmpDir
	if output, err := checkoutCmd.CombinedOutput(); err != nil {
		sendEvent(ImportProgress{Stage: "error", Message: fmt.Sprintf("设置稀疏检出失败: %s", string(output))})
		return
	}

	// 阶段 3：扫描 .md 文件
	sendEvent(ImportProgress{Stage: "scanning", Message: "正在扫描文件..."})

	type mdFileInfo struct {
		Path    string // 绝对路径
		RelPath string // 相对于仓库根的路径
		Dir     string // 文档所在目录
	}
	var mdFiles []mdFileInfo
	filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(tmpDir, path)
		if strings.HasPrefix(rel, ".git") || strings.Contains(rel, "/.git/") {
			return nil
		}
		for _, part := range strings.Split(rel, string(os.PathSeparator)) {
			if strings.HasPrefix(part, ".") {
				return nil
			}
		}
		if strings.EqualFold(filepath.Ext(path), ".md") && info.Size() <= 1*1024*1024 {
			dir := filepath.Dir(rel)
			if dir == "." {
				dir = ""
			}
			mdFiles = append(mdFiles, mdFileInfo{Path: path, RelPath: rel, Dir: dir})
		}
		return nil
	})

	if len(mdFiles) == 0 {
		sendEvent(ImportProgress{Stage: "error", Message: "仓库中未找到 .md 文件"})
		return
	}

	sendEvent(ImportProgress{
		Stage:   "scanning",
		Total:   len(mdFiles),
		Message: fmt.Sprintf("扫描到 %d 个 Markdown 文件", len(mdFiles)),
	})

	// 阶段 4：导入文档，同时收集图片引用
	result := &ImportResult{}
	createdCategories := make(map[string]int64)
	imported := 0
	skipped := 0
	seenTitles := make(map[string]bool)

	// 收集所有图片引用（绝对路径 → 原始引用列表）
	type imageUsage struct {
		Originals map[string][]int64 // 原始引用 → 使用该图片的文档ID列表
		AbsPath   string
	}
	allImageRefs := make(map[string]*imageUsage) // absPath → imageUsage

	for i, mf := range mdFiles {
		sendEvent(ImportProgress{
			Stage:       "importing",
			Processed:   i + 1,
			Total:       len(mdFiles),
			CurrentFile: mf.RelPath,
		})

		content, err := os.ReadFile(mf.Path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("读取 %s 失败: %v", mf.RelPath, err))
			skipped++
			continue
		}

		contentStr := string(content)

		// 清理语雀来源链接
		contentStr = utils.CleanYuqueLinks(contentStr)

		// 解析目录结构 → 分类
		parts := strings.Split(mf.Dir, string(os.PathSeparator))
		var cleanParts []string
		for _, p := range parts {
			if p != "." && p != "" {
				cleanParts = append(cleanParts, p)
			}
		}

		if req.SkipRoot && len(cleanParts) > 0 {
			cleanParts = cleanParts[1:]
		}

		var categoryID *int64
		if len(cleanParts) > 0 {
			catID, err := ensureCategoryPath(cleanParts, createdCategories)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("创建分类 %s 失败: %v", mf.Dir, err))
			} else {
				categoryID = &catID
			}
		}

		title := strings.TrimSuffix(filepath.Base(mf.Path), ".md")

		if seenTitles[title] {
			skipped++
			result.DocsSkipped++
			continue
		}
		existingDoc, _ := findDocumentByTitle(title)
		if existingDoc != nil {
			skipped++
			result.DocsSkipped++
			seenTitles[title] = true
			continue
		}
		seenTitles[title] = true

		doc := &models.Document{
			Title:       title,
			Content:     contentStr,
			CategoryID:  categoryID,
			AuthorID:    1,
			IsPublished: true,
		}
		docID, err := models.CreateDocument(doc)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("创建文档 %s 失败: %v", title, err))
			skipped++
			continue
		}

		imported++
		result.DocsCreated++

		// 收集此文档中的图片引用
		for _, ref := range extractImageRefs(contentStr, mf.Dir) {
			if _, exists := allImageRefs[ref.AbsPath]; !exists {
				allImageRefs[ref.AbsPath] = &imageUsage{
					Originals: make(map[string][]int64),
				}
			}
			allImageRefs[ref.AbsPath].Originals[ref.Original] = append(
				allImageRefs[ref.AbsPath].Originals[ref.Original], docID)
		}

		if imported%50 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	result.CategoriesCreated = len(createdCategories)

	// 阶段 5：通过 GitHub raw URL 下载图片
	imagesImported := 0
	// absPath → 本地上传路径
	imageLocalMap := make(map[string]string)

	if len(allImageRefs) > 0 {
		sendEvent(ImportProgress{
			Stage:   "importing",
			Total:   len(allImageRefs),
			Message: fmt.Sprintf("正在下载 %d 个引用图片...", len(allImageRefs)),
		})

		uploadDir := os.Getenv("UPLOAD_DIR")
		if uploadDir == "" {
			uploadDir = "/app/uploads"
		}

		rawBase := githubURLToRaw(req.URL) + "/" + req.Branch
		client := &http.Client{Timeout: 30 * time.Second}

		i := 0
		for absPath := range allImageRefs {
			i++
			sendEvent(ImportProgress{
				Stage:       "importing",
				Processed:   i,
				Total:       len(allImageRefs),
				CurrentFile: absPath,
			})

			rawURL := rawBase + "/" + absPath

			resp, err := client.Get(rawURL)
			if err != nil || resp.StatusCode != 200 {
				if resp != nil {
					resp.Body.Close()
				}
				continue
			}

			if resp.ContentLength > 5*1024*1024 {
				resp.Body.Close()
				continue
			}

			imgData, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil || len(imgData) > 5*1024*1024 || len(imgData) < 100 {
				continue
			}

			ext := strings.ToLower(filepath.Ext(absPath))
			if ext == "" {
				ext = ".png"
			}
			newName := fmt.Sprintf("gh_%s_%s%s", uuid.New().String()[:8], sanitizeFilename(filepath.Base(absPath)), ext)
			newPath := filepath.Join(uploadDir, newName)

			if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
				continue
			}
			if err := os.WriteFile(newPath, imgData, 0644); err != nil {
				continue
			}

			imageLocalMap[absPath] = "/uploads/" + newName
			imagesImported++

			if imagesImported%10 == 0 {
				time.Sleep(200 * time.Millisecond)
			}
		}
	}

	result.ImagesImported = imagesImported

	// 阶段 6：更新文档中的图片引用（按文档批量处理）
	if imagesImported > 0 {
		sendEvent(ImportProgress{Stage: "importing", Message: "正在更新文档中的图片引用..."})

		// 构建：docID → [(oldRef, newRef)]
		docUpdates := make(map[int64][][2]string)

		for absPath, usage := range allImageRefs {
			newPath, ok := imageLocalMap[absPath]
			if !ok {
				continue
			}
			for origRef, docIDs := range usage.Originals {
				for _, docID := range docIDs {
					docUpdates[docID] = append(docUpdates[docID], [2]string{origRef, newPath})
				}
			}
		}

		// 按文档更新
		updatedDocs := 0
		for docID, replacements := range docUpdates {
			doc, err := models.GetDocumentByID(docID)
			if err != nil || doc == nil {
				continue
			}
			content := doc.Content
			changed := false
			for _, r := range replacements {
				oldRef := fmt.Sprintf("](%s)", r[0])
				newRef := fmt.Sprintf("](%s)", r[1])
				if strings.Contains(content, oldRef) {
					content = strings.ReplaceAll(content, oldRef, newRef)
					changed = true
				}
				// 也处理带 ./
				oldRef2 := fmt.Sprintf("](./%s)", r[0])
				if strings.Contains(content, oldRef2) {
					content = strings.ReplaceAll(content, oldRef2, newRef)
					changed = true
				}
			}
			if changed {
				models.UpdateDocumentContent(docID, content)
				updatedDocs++
			}
		}
	}

	// 发送结果
	resultJSON, _ := json.Marshal(result)
	sendEvent(ImportProgress{Stage: "result", Message: string(resultJSON)})
	sendEvent(ImportProgress{Stage: "done"})

	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "import_github",
		TargetType: "import",
		Detail:     fmt.Sprintf("url:%s, branch:%s, imported:%d, skipped:%d, images:%d", req.URL, req.Branch, imported, skipped, imagesImported),
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
