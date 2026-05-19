package handlers

import (
	"fmt"
	"litewiki/models"
	"litewiki/services"
	"litewiki/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminImportTask 导入任务管理处理器
type AdminImportTask struct {
	UploadDir string
}

// List 获取导入任务列表
func (h *AdminImportTask) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	tasks, total, err := models.ListImportTasks(page, size)
	if err != nil {
		utils.Error(c, 500, 50000, "获取任务列表失败")
		return
	}

	utils.Success(c, utils.PageResult{
		List:  tasks,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// Get 获取单个任务详情
func (h *AdminImportTask) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的任务ID")
		return
	}

	task, err := models.GetImportTaskByID(id)
	if err != nil {
		utils.Error(c, 500, 50000, "获取任务失败")
		return
	}
	if task == nil {
		utils.Error(c, 404, 40400, "任务不存在")
		return
	}

	utils.Success(c, task)
}

// Retry 重试失败任务
func (h *AdminImportTask) Retry(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的任务ID")
		return
	}

	task, err := models.GetImportTaskByID(id)
	if err != nil || task == nil {
		utils.Error(c, 404, 40400, "任务不存在")
		return
	}

	if task.Status != "failed" && task.Status != "cancelled" {
		utils.Error(c, 400, 40002, "只能重试失败或已取消的任务")
		return
	}

	// 重置任务状态
	if err := models.UpdateImportTaskStatus(id, "pending"); err != nil {
		utils.Error(c, 500, 50000, "重置任务失败")
		return
	}
	models.UpdateImportTaskProgress(id, 0, 0, 0, 0, 0, 0)

	// 重新加入队列
	services.GlobalQueue.Enqueue(id)

	utils.Success(c, gin.H{"message": "任务已重新加入队列"})
}

// Delete 删除任务记录
func (h *AdminImportTask) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的任务ID")
		return
	}

	task, err := models.GetImportTaskByID(id)
	if err != nil || task == nil {
		utils.Error(c, 404, 40400, "任务不存在")
		return
	}

	if task.Status == "running" {
		utils.Error(c, 400, 40003, "不能删除正在运行的任务，请先取消")
		return
	}

	if err := models.DeleteImportTask(id); err != nil {
		utils.Error(c, 500, 50000, "删除任务失败")
		return
	}

	utils.Success(c, gin.H{"message": "任务已删除"})
}

// Cancel 取消等待中的任务
func (h *AdminImportTask) Cancel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, 40001, "无效的任务ID")
		return
	}

	if err := services.GlobalQueue.CancelTask(id); err != nil {
		utils.Error(c, 400, 40002, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "任务已取消"})
}

// ImportAsync 异步 ZIP 导入
func (h *AdminImportTask) ImportAsync(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, 400, 40001, "请选择文件")
		return
	}

	if file.Size > 500*1024*1024 {
		utils.Error(c, 400, 40002, "文件大小不能超过 500MB")
		return
	}

	if !strings.HasSuffix(strings.ToLower(file.Filename), ".zip") {
		utils.Error(c, 400, 40003, "请上传 .zip 文件")
		return
	}

	// 保存到临时目录
	tmpDir := filepath.Join(os.TempDir(), "litewiki_imports")
	os.MkdirAll(tmpDir, 0755)
	tmpPath := filepath.Join(tmpDir, fmt.Sprintf("%s_%s", strconv.FormatInt(int64(file.Size), 36), file.Filename))

	if err := c.SaveUploadedFile(file, tmpPath); err != nil {
		utils.Error(c, 500, 50001, "保存文件失败")
		return
	}

	// 创建任务
	taskID, err := models.CreateImportTask("zip", tmpPath)
	if err != nil {
		utils.Error(c, 500, 50002, "创建任务失败")
		return
	}

	// 加入队列
	services.GlobalQueue.Enqueue(taskID)

	// 审计
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "import_zip_async_create",
		TargetType: "import_task",
		TargetID:   taskID,
		Detail:     fmt.Sprintf("file:%s", file.Filename),
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, gin.H{"task_id": taskID, "message": "导入任务已创建"})
}

// ImportBatch 批量 ZIP 导入：接收多个文件，一次性创建所有任务
func (h *AdminImportTask) ImportBatch(c *gin.Context) {
	// 解析 multipart form
	if err := c.Request.ParseMultipartForm(500 << 20); err != nil {
		utils.Error(c, 400, 40001, "解析表单失败")
		return
	}

	files := c.Request.MultipartForm.File["files"]
	if len(files) == 0 {
		utils.Error(c, 400, 40002, "请选择文件")
		return
	}

	tmpDir := filepath.Join(os.TempDir(), "litewiki_imports")
	os.MkdirAll(tmpDir, 0755)

	var taskIDs []int64
	var filenames []string

	for _, fileHeader := range files {
		// 验证
		if fileHeader.Size > 500*1024*1024 {
			utils.Error(c, 400, 40003, fmt.Sprintf("文件 %s 大小超过 500MB", fileHeader.Filename))
			return
		}
		if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".zip") {
			utils.Error(c, 400, 40004, fmt.Sprintf("文件 %s 不是 .zip 文件", fileHeader.Filename))
			return
		}

		// 保存到临时目录
		tmpPath := filepath.Join(tmpDir, fmt.Sprintf("%s_%s", strconv.FormatInt(int64(fileHeader.Size), 36), fileHeader.Filename))
		if err := c.SaveUploadedFile(fileHeader, tmpPath); err != nil {
			utils.Error(c, 500, 50001, fmt.Sprintf("保存文件 %s 失败", fileHeader.Filename))
			return
		}

		// 创建任务
		taskID, err := models.CreateImportTask("zip", tmpPath)
		if err != nil {
			utils.Error(c, 500, 50002, fmt.Sprintf("为文件 %s 创建任务失败", fileHeader.Filename))
			return
		}

		// 加入队列
		services.GlobalQueue.Enqueue(taskID)

		taskIDs = append(taskIDs, taskID)
		filenames = append(filenames, fileHeader.Filename)
	}

	// 审计
	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "import_zip_batch_create",
		TargetType: "import_task",
		Detail:     fmt.Sprintf("count:%d, files:%s", len(taskIDs), strings.Join(filenames, ",")),
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, gin.H{
		"task_ids": taskIDs,
		"count":    len(taskIDs),
		"message":  fmt.Sprintf("已创建 %d 个导入任务", len(taskIDs)),
	})
}

// GitHubImportAsync 异步 GitHub 导入
func (h *AdminImportTask) GitHubImportAsync(c *gin.Context) {
	var req struct {
		URL      string `json:"url" binding:"required"`
		Branch   string `json:"branch"`
		SkipRoot bool   `json:"skip_root"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, 40001, "请求参数错误")
		return
	}

	// source 格式: "url|branch|skipRoot"
	source := fmt.Sprintf("%s|%s|%v", req.URL, req.Branch, req.SkipRoot)

	taskID, err := models.CreateImportTask("github", source)
	if err != nil {
		utils.Error(c, 500, 50002, "创建任务失败")
		return
	}

	services.GlobalQueue.Enqueue(taskID)

	go models.InsertAuditLog(&models.AuditLog{
		UserID:     c.GetInt64("user_id"),
		Username:   c.GetString("username"),
		Action:     "import_github_async_create",
		TargetType: "import_task",
		TargetID:   taskID,
		Detail:     fmt.Sprintf("url:%s", req.URL),
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	utils.Success(c, gin.H{"task_id": taskID, "message": "导入任务已创建"})
}
