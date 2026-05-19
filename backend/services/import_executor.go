package services

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"litewiki/models"
	"litewiki/utils"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// progressCallback 进度回调
type progressCallback func(progress int, totalDocs, importedDocs, updatedDocs, skippedDocs, errorCount int)

// executeZipImport 执行 ZIP 导入任务
func executeZipImport(ctx context.Context, task *models.ImportTask) {
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "/app/uploads"
	}

	// 读取 ZIP 文件
	zipPath := task.Source
	zipData, err := os.ReadFile(zipPath)
	if err != nil {
		log.Error().Err(err).Int64("taskID", task.ID).Msg("读取 ZIP 文件失败")
		models.UpdateImportTaskStatus(task.ID, "failed")
		models.UpdateImportTaskResult(task.ID, nil, []string{fmt.Sprintf("读取文件失败: %v", err)})
		return
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		models.UpdateImportTaskStatus(task.ID, "failed")
		models.UpdateImportTaskResult(task.ID, nil, []string{"无效的 ZIP 文件"})
		return
	}

	// 扫描文件
	type zipEntry struct {
		Path string
		File *zip.File
	}
	var mdFiles []zipEntry
	var imgFiles []zipEntry
	ignoredPrefixes := []string{"__MACOSX/", ".DS_Store", "._"}

	for _, f := range zipReader.File {
		fName := decodeZipFilename(f.Name, f.NonUTF8)
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
			if f.UncompressedSize64 > 1*1024*1024 {
				continue
			}
			mdFiles = append(mdFiles, zipEntry{Path: fName, File: f})
		case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg":
			imgFiles = append(imgFiles, zipEntry{Path: fName, File: f})
		}
	}

	totalFiles := len(mdFiles) + len(imgFiles)
	models.UpdateImportTaskProgress(task.ID, 0, len(mdFiles), 0, 0, 0, 0)

	// 导入图片
	imageURLMap := make(map[string]string)
	for _, entry := range imgFiles {
		select {
		case <-ctx.Done():
			models.UpdateImportTaskStatus(task.ID, "cancelled")
			return
		default:
		}

		rc, err := entry.File.Open()
		if err != nil {
			continue
		}
		ext := filepath.Ext(entry.Path)
		filename := time.Now().Format("20060102") + "/" + uuid.New().String() + ext
		dst := filepath.Join(uploadDir, filename)
		if err := utils.EnsureDir(filepath.Dir(dst)); err != nil {
			rc.Close()
			continue
		}
		data, _ := io.ReadAll(rc)
		rc.Close()
		if err := utils.WriteFile(dst, data); err != nil {
			continue
		}
		imageURLMap[entry.Path] = "/uploads/" + filename
		imageURLMap[filepath.Base(entry.Path)] = "/uploads/" + filename
	}

	// 导入文档
	createdCategories := make(map[string]int64)
	importedDocs := 0
	updatedDocs := 0
	skippedDocs := 0
	var errors []string

	for i, entry := range mdFiles {
		select {
		case <-ctx.Done():
			models.UpdateImportTaskStatus(task.ID, "cancelled")
			return
		default:
		}

		progress := 0
		if totalFiles > 0 {
			progress = int(float64(i+1+len(imgFiles)) / float64(totalFiles) * 100)
		}

		rc, err := entry.File.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("打开 %s 失败: %v", entry.Path, err))
			skippedDocs++
			models.UpdateImportTaskProgress(task.ID, progress, len(mdFiles), importedDocs, updatedDocs, skippedDocs, len(errors))
			continue
		}
		content, _ := io.ReadAll(rc)
		rc.Close()

		contentStr := detectAndConvertEncoding(content)
		contentStr = utils.CleanYuqueLinks(contentStr)
		contentStr = replaceImagePaths(contentStr, imageURLMap, entry.Path)

		dir := filepath.Dir(entry.Path)
		parts := strings.Split(dir, "/")
		if len(parts) > 0 && parts[0] == "" {
			parts = parts[1:]
		}

		var categoryID *int64
		if len(parts) > 0 && parts[0] != "." {
			catID, err := ensureCategoryPathLocal(parts, createdCategories)
			if err != nil {
				errors = append(errors, fmt.Sprintf("创建分类 %s 失败: %v", dir, err))
			} else {
				categoryID = &catID
			}
		}

		title := strings.TrimSuffix(filepath.Base(entry.Path), ".md")
		existingDoc, _ := models.GetDocumentByTitle(title)
		if existingDoc != nil {
			existingDoc.Content = contentStr
			if categoryID != nil {
				existingDoc.CategoryID = categoryID
			}
			if err := models.UpdateDocument(existingDoc); err != nil {
				errors = append(errors, fmt.Sprintf("更新文档 %s 失败: %v", title, err))
			} else {
				updatedDocs++
			}
		} else {
			doc := &models.Document{
				Title: title, Content: contentStr, CategoryID: categoryID,
				AuthorID: 1, IsPublished: true,
			}
			if _, err := models.CreateDocument(doc); err != nil {
				errors = append(errors, fmt.Sprintf("创建文档 %s 失败: %v", title, err))
			} else {
				importedDocs++
			}
		}

		models.UpdateImportTaskProgress(task.ID, progress, len(mdFiles), importedDocs, updatedDocs, skippedDocs, len(errors))
	}

	// 完成
	result := map[string]int{
		"categories_created": len(createdCategories),
		"docs_created":       importedDocs,
		"docs_updated":       updatedDocs,
		"docs_skipped":       skippedDocs,
		"images_imported":    len(imageURLMap) / 2, // 每个图片存了两个key
	}
	models.UpdateImportTaskResult(task.ID, result, errors)
	models.UpdateImportTaskStatus(task.ID, "completed")

	// 审计日志
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     1,
		Username:   "system",
		Action:     "import_zip_async",
		TargetType: "import",
		Detail:     fmt.Sprintf("task:%d, imported:%d, updated:%d", task.ID, importedDocs, updatedDocs),
	})

	// 清理临时文件
	os.Remove(zipPath)
}

// executeGitHubImport 执行 GitHub 导入任务
func executeGitHubImport(ctx context.Context, task *models.ImportTask) {
	// source 格式: "url|branch|skipRoot"
	parts := strings.SplitN(task.Source, "|", 3)
	if len(parts) < 1 {
		models.UpdateImportTaskStatus(task.ID, "failed")
		models.UpdateImportTaskResult(task.ID, nil, []string{"无效的 source 格式"})
		return
	}

	repoURL := parts[0]
	branch := "main"
	skipRoot := false
	if len(parts) > 1 && parts[1] != "" {
		branch = parts[1]
	}
	if len(parts) > 2 && parts[2] == "true" {
		skipRoot = true
	}

	// 自动检测分支
	if branch == "main" {
		out, err := exec.Command("git", "ls-remote", "--symref", repoURL, "HEAD").CombinedOutput()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				if strings.HasPrefix(line, "ref: refs/heads/") {
					p := strings.SplitN(line, "\t", 2)
					b := strings.TrimPrefix(p[0], "ref: refs/heads/")
					if b != "" {
						branch = b
					}
				}
			}
		}
	}

	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("github_import_%s", uuid.New().String()[:8]))
	defer os.RemoveAll(tmpDir)

	// Clone
	select {
	case <-ctx.Done():
		models.UpdateImportTaskStatus(task.ID, "cancelled")
		return
	default:
	}

	cloneArgs := []string{
		"clone", "--filter=blob:none", "--sparse", "--depth", "1",
		"--single-branch", "--branch", branch, repoURL, tmpDir,
	}
	cloneCmd := exec.Command("git", cloneArgs...)
	cloneCmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cloneCmd.CombinedOutput()
	if err != nil {
		errMsg := string(output)
		if strings.Contains(errMsg, "not found") {
			models.UpdateImportTaskStatus(task.ID, "failed")
			models.UpdateImportTaskResult(task.ID, nil, []string{"仓库不存在"})
		} else {
			models.UpdateImportTaskStatus(task.ID, "failed")
			models.UpdateImportTaskResult(task.ID, nil, []string{fmt.Sprintf("克隆失败: %s", truncateStr(errMsg, 200))})
		}
		return
	}

	// sparse-checkout
	initCmd := exec.Command("git", "sparse-checkout", "init", "--no-cone")
	initCmd.Dir = tmpDir
	initCmd.CombinedOutput()

	checkoutCmd := exec.Command("git", "sparse-checkout", "set", "*.md")
	checkoutCmd.Dir = tmpDir
	checkoutCmd.CombinedOutput()

	// 扫描 md 文件
	type mdFileInfo struct {
		Path, RelPath, Dir string
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
		models.UpdateImportTaskStatus(task.ID, "failed")
		models.UpdateImportTaskResult(task.ID, nil, []string{"未找到 .md 文件"})
		return
	}

	models.UpdateImportTaskProgress(task.ID, 0, len(mdFiles), 0, 0, 0, 0)

	// 导入文档
	createdCategories := make(map[string]int64)
	importedDocs := 0
	skippedDocs := 0
	var errors []string
	seenTitles := make(map[string]bool)

	type imageUsage struct {
		Originals map[string][]int64
		AbsPath   string
	}
	allImageRefs := make(map[string]*imageUsage)

	for i, mf := range mdFiles {
		select {
		case <-ctx.Done():
			models.UpdateImportTaskStatus(task.ID, "cancelled")
			return
		default:
		}

		progress := int(float64(i+1) / float64(len(mdFiles)) * 80) // 80% for doc import

		content, err := os.ReadFile(mf.Path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("读取 %s 失败: %v", mf.RelPath, err))
			skippedDocs++
			models.UpdateImportTaskProgress(task.ID, progress, len(mdFiles), importedDocs, 0, skippedDocs, len(errors))
			continue
		}

		contentStr := string(content)
		contentStr = utils.CleanYuqueLinks(contentStr)

		dirParts := strings.Split(mf.Dir, string(os.PathSeparator))
		var cleanParts []string
		for _, p := range dirParts {
			if p != "." && p != "" {
				cleanParts = append(cleanParts, p)
			}
		}
		if skipRoot && len(cleanParts) > 0 {
			cleanParts = cleanParts[1:]
		}

		var categoryID *int64
		if len(cleanParts) > 0 {
			catID, err := ensureCategoryPathLocal(cleanParts, createdCategories)
			if err != nil {
				errors = append(errors, fmt.Sprintf("创建分类 %s 失败: %v", mf.Dir, err))
			} else {
				categoryID = &catID
			}
		}

		title := strings.TrimSuffix(filepath.Base(mf.Path), ".md")
		if seenTitles[title] {
			skippedDocs++
			models.UpdateImportTaskProgress(task.ID, progress, len(mdFiles), importedDocs, 0, skippedDocs, len(errors))
			continue
		}
		existingDoc, _ := models.GetDocumentByTitle(title)
		if existingDoc != nil {
			skippedDocs++
			seenTitles[title] = true
			models.UpdateImportTaskProgress(task.ID, progress, len(mdFiles), importedDocs, 0, skippedDocs, len(errors))
			continue
		}
		seenTitles[title] = true

		doc := &models.Document{
			Title: title, Content: contentStr, CategoryID: categoryID,
			AuthorID: 1, IsPublished: true,
		}
		docID, err := models.CreateDocument(doc)
		if err != nil {
			errors = append(errors, fmt.Sprintf("创建文档 %s 失败: %v", title, err))
			skippedDocs++
			models.UpdateImportTaskProgress(task.ID, progress, len(mdFiles), importedDocs, 0, skippedDocs, len(errors))
			continue
		}

		importedDocs++
		for _, ref := range extractImageRefsLocal(contentStr, mf.Dir) {
			if _, exists := allImageRefs[ref.AbsPath]; !exists {
				allImageRefs[ref.AbsPath] = &imageUsage{Originals: make(map[string][]int64)}
			}
			allImageRefs[ref.AbsPath].Originals[ref.Original] = append(
				allImageRefs[ref.AbsPath].Originals[ref.Original], docID)
		}

		models.UpdateImportTaskProgress(task.ID, progress, len(mdFiles), importedDocs, 0, skippedDocs, len(errors))
	}

	// 下载图片
	imagesImported := 0
	imageLocalMap := make(map[string]string)
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "/app/uploads"
	}

	if len(allImageRefs) > 0 {
		rawBase := githubURLToRawLocal(repoURL) + "/" + branch
		client := &http.Client{Timeout: 30 * time.Second}

		i := 0
		for absPath := range allImageRefs {
			select {
			case <-ctx.Done():
				models.UpdateImportTaskStatus(task.ID, "cancelled")
				return
			default:
			}
			i++
			progress := 80 + int(float64(i)/float64(len(allImageRefs))*19) // 80-99%

			rawURL := rawBase + "/" + absPath
			resp, err := client.Get(rawURL)
			if err != nil || resp.StatusCode != 200 {
				if resp != nil {
					resp.Body.Close()
				}
				models.UpdateImportTaskProgress(task.ID, progress, len(mdFiles), importedDocs, 0, skippedDocs, len(errors))
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
			newName := fmt.Sprintf("gh_%s_%s%s", uuid.New().String()[:8], sanitizeFilenameLocal(filepath.Base(absPath)), ext)
			newPath := filepath.Join(uploadDir, newName)
			if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
				continue
			}
			if err := os.WriteFile(newPath, imgData, 0644); err != nil {
				continue
			}
			imageLocalMap[absPath] = "/uploads/" + newName
			imagesImported++

			models.UpdateImportTaskProgress(task.ID, progress, len(mdFiles), importedDocs, 0, skippedDocs, len(errors))
		}
	}

	// 更新文档中的图片引用
	if imagesImported > 0 {
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
				oldRef2 := fmt.Sprintf("](./%s)", r[0])
				if strings.Contains(content, oldRef2) {
					content = strings.ReplaceAll(content, oldRef2, newRef)
					changed = true
				}
			}
			if changed {
				models.UpdateDocumentContent(docID, content)
			}
		}
	}

	// 完成
	result := map[string]int{
		"categories_created": len(createdCategories),
		"docs_created":       importedDocs,
		"docs_skipped":       skippedDocs,
		"images_imported":    imagesImported,
	}
	models.UpdateImportTaskResult(task.ID, result, errors)
	models.UpdateImportTaskStatus(task.ID, "completed")

	go models.InsertAuditLog(&models.AuditLog{
		UserID:     1,
		Username:   "system",
		Action:     "import_github_async",
		TargetType: "import",
		Detail:     fmt.Sprintf("task:%d, url:%s, imported:%d, skipped:%d", task.ID, repoURL, importedDocs, skippedDocs),
	})
}

// --- 以下是从 import.go 和 github_import.go 提取的辅助函数 ---

func ensureCategoryPathLocal(parts []string, cache map[string]int64) (int64, error) {
	if len(parts) == 0 {
		return 0, fmt.Errorf("空路径")
	}
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
		cat, err := FindCategoryByNameAndParent(part, parentID)
		if err != nil || cat == nil {
			newCat, err := CreateCategory(&CreateCategoryRequest{Name: part, ParentID: parentID})
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

func detectAndConvertEncoding(data []byte) string {
	if utf8.Valid(data) {
		return string(data)
	}
	if isLikelyGBK(data) {
		reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder())
		converted, err := io.ReadAll(reader)
		if err == nil && utf8.Valid(converted) {
			return string(converted)
		}
	}
	reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GB18030.NewDecoder())
	converted, err := io.ReadAll(reader)
	if err == nil && utf8.Valid(converted) {
		return string(converted)
	}
	return string(data)
}

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

func decodeZipFilename(name string, nonUTF8 bool) string {
	if !nonUTF8 && utf8.ValidString(name) {
		return name
	}
	reader := transform.NewReader(bytes.NewReader([]byte(name)), simplifiedchinese.GBK.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err == nil && utf8.Valid(decoded) {
		return string(decoded)
	}
	reader = transform.NewReader(bytes.NewReader([]byte(name)), simplifiedchinese.GB18030.NewDecoder())
	decoded, err = io.ReadAll(reader)
	if err == nil && utf8.Valid(decoded) {
		return string(decoded)
	}
	return name
}

func replaceImagePaths(content string, urlMap map[string]string, mdPath string) string {
	result := content
	mdDir := filepath.Dir(mdPath)
	for origPath, newURL := range urlMap {
		if strings.HasPrefix(origPath, mdDir+"/") {
			relPath := strings.TrimPrefix(origPath, mdDir+"/")
			result = strings.ReplaceAll(result, "(./"+relPath+")", "("+newURL+")")
			result = strings.ReplaceAll(result, "("+relPath+")", "("+newURL+")")
		}
		baseName := filepath.Base(origPath)
		result = strings.ReplaceAll(result, "(./img/"+baseName+")", "("+newURL+")")
		result = strings.ReplaceAll(result, "(img/"+baseName+")", "("+newURL+")")
		result = strings.ReplaceAll(result, "("+baseName+")", "("+newURL+")")
	}
	return result
}

type imageRef struct {
	AbsPath  string
	Original string
}

func extractImageRefs(content string, dir string) []imageRef {
	var refs []imageRef
	seen := make(map[string]bool)
	re := regexp.MustCompile(`!\[([^\]]*)\]\(([^)\s]+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			origPath := m[2]
			if strings.HasPrefix(origPath, "http://") || strings.HasPrefix(origPath, "https://") || strings.HasPrefix(origPath, "#") {
				continue
			}
			absPath := resolveRepoPathLocal(origPath, dir)
			key := absPath + "|" + origPath
			if !seen[key] {
				seen[key] = true
				refs = append(refs, imageRef{AbsPath: absPath, Original: origPath})
			}
		}
	}
	imgRe := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`)
	imgMatches := imgRe.FindAllStringSubmatch(content, -1)
	for _, m := range imgMatches {
		if len(m) >= 2 {
			origPath := m[1]
			if strings.HasPrefix(origPath, "http://") || strings.HasPrefix(origPath, "https://") || strings.HasPrefix(origPath, "#") {
				continue
			}
			absPath := resolveRepoPathLocal(origPath, dir)
			key := absPath + "|" + origPath
			if !seen[key] {
				seen[key] = true
				refs = append(refs, imageRef{AbsPath: absPath, Original: origPath})
			}
		}
	}
	return refs
}

// extractImageRefsLocal 是 extractImageRefs 的别名（避免与 github_import.go 中的同名函数冲突）
func extractImageRefsLocal(content string, dir string) []imageRef {
	return extractImageRefs(content, dir)
}

func resolveRepoPathLocal(relPath, docDir string) string {
	if docDir == "" || docDir == "." {
		return filepath.Clean(relPath)
	}
	return filepath.Clean(filepath.Join(docDir, relPath))
}

func githubURLToRawLocal(url string) string {
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimPrefix(url, "https://github.com/")
	return "https://raw.githubusercontent.com/" + url
}

func sanitizeFilenameLocal(name string) string {
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

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
