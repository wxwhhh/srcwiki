package models

import (
	"database/sql"
	"encoding/json"
	"litewiki/database"
	"time"
)

// ImportTask 导入任务模型
type ImportTask struct {
	ID            int64      `json:"id"`
	Type          string     `json:"type"`           // "zip" 或 "github"
	Status        string     `json:"status"`         // pending/running/completed/failed/cancelled
	Source        string     `json:"source"`          // ZIP文件名 或 GitHub URL
	Progress      int        `json:"progress"`        // 0-100
	TotalDocs     int        `json:"total_docs"`
	ImportedDocs  int        `json:"imported_docs"`
	UpdatedDocs   int        `json:"updated_docs"`
	SkippedDocs   int        `json:"skipped_docs"`
	ErrorCount    int        `json:"error_count"`
	Errors        string     `json:"errors"`          // JSON数组
	Result        string     `json:"result"`          // JSON
	CreatedAt     time.Time  `json:"created_at"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
}

// ImportTaskListItem 任务列表项
type ImportTaskListItem struct {
	ID           int64      `json:"id"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	Source       string     `json:"source"`
	Progress     int        `json:"progress"`
	TotalDocs    int        `json:"total_docs"`
	ImportedDocs int        `json:"imported_docs"`
	UpdatedDocs  int        `json:"updated_docs"`
	SkippedDocs  int        `json:"skipped_docs"`
	ErrorCount   int        `json:"error_count"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
}

// CreateImportTask 创建导入任务
func CreateImportTask(taskType, source string) (int64, error) {
	result, err := database.DB.Exec(
		"INSERT INTO import_tasks (type, status, source) VALUES (?, 'pending', ?)",
		taskType, source,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetImportTaskByID 根据ID获取任务
func GetImportTaskByID(id int64) (*ImportTask, error) {
	t := &ImportTask{}
	err := database.DB.QueryRow(
		`SELECT id, type, status, source, progress, total_docs, imported_docs, updated_docs,
		skipped_docs, error_count, COALESCE(errors, ''), COALESCE(result, ''), created_at, started_at, finished_at
		FROM import_tasks WHERE id = ?`, id,
	).Scan(&t.ID, &t.Type, &t.Status, &t.Source, &t.Progress, &t.TotalDocs,
		&t.ImportedDocs, &t.UpdatedDocs, &t.SkippedDocs, &t.ErrorCount,
		&t.Errors, &t.Result, &t.CreatedAt, &t.StartedAt, &t.FinishedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

// ListImportTasks 获取任务列表
func ListImportTasks(page, size int) ([]ImportTaskListItem, int64, error) {
	var total int64
	err := database.DB.QueryRow("SELECT COUNT(*) FROM import_tasks").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	rows, err := database.DB.Query(
		`SELECT id, type, status, source, progress, total_docs, imported_docs, updated_docs,
		skipped_docs, error_count, created_at, started_at, finished_at
		FROM import_tasks ORDER BY id DESC LIMIT ? OFFSET ?`, size, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []ImportTaskListItem
	for rows.Next() {
		var t ImportTaskListItem
		if err := rows.Scan(&t.ID, &t.Type, &t.Status, &t.Source, &t.Progress, &t.TotalDocs,
			&t.ImportedDocs, &t.UpdatedDocs, &t.SkippedDocs, &t.ErrorCount,
			&t.CreatedAt, &t.StartedAt, &t.FinishedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, nil
}

// UpdateImportTaskStatus 更新任务状态
func UpdateImportTaskStatus(id int64, status string) error {
	if status == "running" {
		_, err := database.DB.Exec(
			"UPDATE import_tasks SET status = ?, started_at = CURRENT_TIMESTAMP WHERE id = ?",
			status, id,
		)
		return err
	}
	if status == "completed" || status == "failed" || status == "cancelled" {
		_, err := database.DB.Exec(
			"UPDATE import_tasks SET status = ?, finished_at = CURRENT_TIMESTAMP WHERE id = ?",
			status, id,
		)
		return err
	}
	_, err := database.DB.Exec("UPDATE import_tasks SET status = ? WHERE id = ?", status, id)
	return err
}

// UpdateImportTaskProgress 更新任务进度
func UpdateImportTaskProgress(id int64, progress, totalDocs, importedDocs, updatedDocs, skippedDocs, errorCount int) error {
	_, err := database.DB.Exec(
		`UPDATE import_tasks SET progress = ?, total_docs = ?, imported_docs = ?,
		updated_docs = ?, skipped_docs = ?, error_count = ? WHERE id = ?`,
		progress, totalDocs, importedDocs, updatedDocs, skippedDocs, errorCount, id,
	)
	return err
}

// UpdateImportTaskResult 更新任务结果
func UpdateImportTaskResult(id int64, result interface{}, errors []string) error {
	resultJSON, _ := json.Marshal(result)
	errorsJSON, _ := json.Marshal(errors)
	_, err := database.DB.Exec(
		"UPDATE import_tasks SET result = ?, errors = ? WHERE id = ?",
		string(resultJSON), string(errorsJSON), id,
	)
	return err
}

// DeleteImportTask 删除任务
func DeleteImportTask(id int64) error {
	_, err := database.DB.Exec("DELETE FROM import_tasks WHERE id = ?", id)
	return err
}

// GetPendingImportTasks 获取等待中的任务（用于恢复）
func GetPendingImportTasks() ([]ImportTask, error) {
	rows, err := database.DB.Query(
		`SELECT id, type, status, source, progress, total_docs, imported_docs, updated_docs,
		skipped_docs, error_count, COALESCE(errors, ''), COALESCE(result, ''), created_at, started_at, finished_at
		FROM import_tasks WHERE status IN ('pending', 'running') ORDER BY id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []ImportTask
	for rows.Next() {
		var t ImportTask
		if err := rows.Scan(&t.ID, &t.Type, &t.Status, &t.Source, &t.Progress, &t.TotalDocs,
			&t.ImportedDocs, &t.UpdatedDocs, &t.SkippedDocs, &t.ErrorCount,
			&t.Errors, &t.Result, &t.CreatedAt, &t.StartedAt, &t.FinishedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// ResetStaleTasks 重置卡在 running 状态的任务（服务重启恢复）
func ResetStaleTasks() error {
	_, err := database.DB.Exec(
		"UPDATE import_tasks SET status = 'pending', started_at = NULL WHERE status = 'running'",
	)
	return err
}
