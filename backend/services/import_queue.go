package services

import (
	"context"
	"fmt"
	"litewiki/models"
	"sync"

	"github.com/rs/zerolog/log"
)

// ImportQueue 导入任务队列管理器
type ImportQueue struct {
	taskChan chan int64
	cancelMap sync.Map // taskID -> context.CancelFunc
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

// GlobalQueue 全局任务队列实例
var GlobalQueue *ImportQueue

// InitImportQueue 初始化导入任务队列
func InitImportQueue() {
	ctx, cancel := context.WithCancel(context.Background())
	GlobalQueue = &ImportQueue{
		taskChan: make(chan int64, 100),
		ctx:      ctx,
		cancel:   cancel,
	}

	// 重置卡在 running 状态的任务（服务重启恢复）
	if err := models.ResetStaleTasks(); err != nil {
		log.Warn().Err(err).Msg("重置 stale 任务失败")
	}

	// 启动消费者
	GlobalQueue.wg.Add(1)
	go GlobalQueue.worker()

	// 恢复等待中的任务
	tasks, err := models.GetPendingImportTasks()
	if err != nil {
		log.Warn().Err(err).Msg("获取待处理任务失败")
		return
	}
	for _, t := range tasks {
		GlobalQueue.Enqueue(t.ID)
	}
	log.Info().Int("pending", len(tasks)).Msg("导入队列初始化完成")
}

// Enqueue 将任务加入队列
func (q *ImportQueue) Enqueue(taskID int64) {
	select {
	case q.taskChan <- taskID:
		log.Info().Int64("taskID", taskID).Msg("任务已加入队列")
	default:
		log.Warn().Int64("taskID", taskID).Msg("队列已满，任务等待")
		go func() {
			q.taskChan <- taskID
		}()
	}
}

// CancelTask 取消任务
func (q *ImportQueue) CancelTask(taskID int64) error {
	if cancelFn, ok := q.cancelMap.Load(taskID); ok {
		cancelFn.(context.CancelFunc)()
		return models.UpdateImportTaskStatus(taskID, "cancelled")
	}
	// 任务还未开始，直接标记取消
	task, err := models.GetImportTaskByID(taskID)
	if err != nil || task == nil {
		return fmt.Errorf("任务不存在")
	}
	if task.Status == "pending" {
		return models.UpdateImportTaskStatus(taskID, "cancelled")
	}
	return fmt.Errorf("任务状态无法取消: %s", task.Status)
}

// worker 消费队列
func (q *ImportQueue) worker() {
	defer q.wg.Done()
	for {
		select {
		case <-q.ctx.Done():
			return
		case taskID := <-q.taskChan:
			q.processTask(taskID)
		}
	}
}

// processTask 处理单个任务
func (q *ImportQueue) processTask(taskID int64) {
	task, err := models.GetImportTaskByID(taskID)
	if err != nil || task == nil {
		log.Error().Int64("taskID", taskID).Msg("获取任务失败")
		return
	}

	// 跳过已取消的任务
	if task.Status == "cancelled" {
		return
	}

	// 创建可取消的 context
	ctx, cancel := context.WithCancel(q.ctx)
	q.cancelMap.Store(taskID, cancel)
	defer func() {
		cancel()
		q.cancelMap.Delete(taskID)
	}()

	// 更新状态为 running
	if err := models.UpdateImportTaskStatus(taskID, "running"); err != nil {
		log.Error().Err(err).Int64("taskID", taskID).Msg("更新任务状态失败")
		return
	}

	log.Info().Int64("taskID", taskID).Str("type", task.Type).Msg("开始处理导入任务")

	switch task.Type {
	case "zip":
		executeZipImport(ctx, task)
	case "github":
		executeGitHubImport(ctx, task)
	default:
		log.Error().Str("type", task.Type).Int64("taskID", taskID).Msg("未知任务类型")
		models.UpdateImportTaskStatus(taskID, "failed")
	}
}

// Shutdown 优雅关闭队列
func (q *ImportQueue) Shutdown() {
	q.cancel()
	q.wg.Wait()
}
