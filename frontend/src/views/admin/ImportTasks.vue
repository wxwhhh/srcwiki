<template>
  <div class="import-tasks-page">
    <div class="page-header">
      <h1 class="page-title">导入任务</h1>
      <el-button @click="loadTasks" :loading="loading" size="small">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <!-- 任务列表 -->
    <div class="tasks-card">
      <div v-if="tasks.length === 0 && !loading" class="empty-state">
        <el-icon :size="48" color="#C7C7CC"><List /></el-icon>
        <p>暂无导入任务</p>
        <router-link to="/admin/dashboard/import" class="link">前往批量导入</router-link>
      </div>

      <div v-else class="task-list">
        <div v-for="task in tasks" :key="task.id" class="task-item" :class="task.status">
          <div class="task-main">
            <div class="task-info">
              <div class="task-header">
                <span class="task-type">{{ task.type === 'zip' ? '📦 ZIP' : '🔗 GitHub' }}</span>
                <span class="task-status" :class="task.status">
                  {{ statusIcon(task.status) }} {{ statusText(task.status) }}
                </span>
                <span class="task-id">#{{ task.id }}</span>
              </div>
              <div class="task-source" :title="task.source">{{ formatSource(task.source) }}</div>
              <div class="task-time">
                创建于 {{ formatTime(task.created_at) }}
                <template v-if="task.finished_at">
                  · 耗时 {{ calcDuration(task.started_at, task.finished_at) }}
                </template>
              </div>
            </div>

            <div class="task-stats" v-if="task.status === 'completed' || task.status === 'failed'">
              <div class="stat" v-if="task.imported_docs > 0">
                <span class="stat-val green">{{ task.imported_docs }}</span>
                <span class="stat-lbl">导入</span>
              </div>
              <div class="stat" v-if="task.updated_docs > 0">
                <span class="stat-val blue">{{ task.updated_docs }}</span>
                <span class="stat-lbl">更新</span>
              </div>
              <div class="stat" v-if="task.skipped_docs > 0">
                <span class="stat-val gray">{{ task.skipped_docs }}</span>
                <span class="stat-lbl">跳过</span>
              </div>
              <div class="stat" v-if="task.error_count > 0">
                <span class="stat-val red">{{ task.error_count }}</span>
                <span class="stat-lbl">错误</span>
              </div>
            </div>

            <div class="task-actions">
              <el-button
                v-if="task.status === 'pending'"
                type="warning"
                size="small"
                plain
                @click="handleCancel(task.id)"
              >
                取消
              </el-button>
              <el-button
                v-if="task.status === 'failed' || task.status === 'cancelled'"
                type="primary"
                size="small"
                plain
                @click="handleRetry(task.id)"
              >
                重试
              </el-button>
              <el-button
                v-if="task.status !== 'running'"
                type="danger"
                size="small"
                plain
                @click="handleDelete(task.id)"
              >
                删除
              </el-button>
            </div>
          </div>

          <!-- 进度条 -->
          <div v-if="task.status === 'running' || task.status === 'pending'" class="task-progress">
            <el-progress
              :percentage="task.progress"
              :status="task.status === 'running' ? undefined : undefined"
              :stroke-width="6"
            />
          </div>

          <!-- 错误信息 -->
          <div v-if="task.status === 'failed' && task.errors" class="task-errors">
            <div v-for="(err, i) in parseErrors(task.errors)" :key="i" class="error-item">
              ⚠️ {{ err }}
            </div>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div v-if="total > pageSize" class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          @current-change="loadTasks"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listImportTasks, retryImportTask, deleteImportTask, cancelImportTask } from '@/api/admin'
import type { ImportTask } from '@/api/admin'

const tasks = ref<ImportTask[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = 20
const total = ref(0)
let pollTimer: ReturnType<typeof setInterval> | null = null

async function loadTasks() {
  loading.value = true
  try {
    const res = await listImportTasks(currentPage.value, pageSize)
    if (res.code === 0 && res.data) {
      tasks.value = res.data.list || []
      total.value = res.data.total
    }
  } catch (e: any) {
    ElMessage.error('加载任务列表失败')
  } finally {
    loading.value = false
  }
}

function hasActiveTasks() {
  return tasks.value.some(t => t.status === 'pending' || t.status === 'running')
}

function startPolling() {
  if (pollTimer) return
  pollTimer = setInterval(() => {
    if (hasActiveTasks()) {
      loadTasks()
    }
  }, 2000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function handleRetry(id: number) {
  try {
    const res = await retryImportTask(id)
    if (res.code === 0) {
      ElMessage.success('任务已重新加入队列')
      loadTasks()
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '重试失败')
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定删除此任务记录？', '确认')
    const res = await deleteImportTask(id)
    if (res.code === 0) {
      ElMessage.success('已删除')
      loadTasks()
    }
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.response?.data?.message || '删除失败')
    }
  }
}

async function handleCancel(id: number) {
  try {
    await ElMessageBox.confirm('确定取消此任务？', '确认')
    const res = await cancelImportTask(id)
    if (res.code === 0) {
      ElMessage.success('任务已取消')
      loadTasks()
    }
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.response?.data?.message || '取消失败')
    }
  }
}

function statusIcon(status: string) {
  switch (status) {
    case 'pending': return '⏳'
    case 'running': return '⏳'
    case 'completed': return '✅'
    case 'failed': return '❌'
    case 'cancelled': return '🚫'
    default: return ''
  }
}

function statusText(status: string) {
  switch (status) {
    case 'pending': return '等待中'
    case 'running': return '进行中'
    case 'completed': return '已完成'
    case 'failed': return '失败'
    case 'cancelled': return '已取消'
    default: return status
  }
}

function formatSource(source: string) {
  if (source.includes('|')) {
    return source.split('|')[0]
  }
  // ZIP: show filename
  const parts = source.split('/')
  return parts[parts.length - 1]
}

function formatTime(t: string) {
  if (!t) return ''
  const d = new Date(t + (t.includes('Z') ? '' : 'Z'))
  return d.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function calcDuration(start: string | null, end: string | null) {
  if (!start || !end) return ''
  const s = new Date(start + (start.includes('Z') ? '' : 'Z')).getTime()
  const e = new Date(end + (end.includes('Z') ? '' : 'Z')).getTime()
  const sec = Math.round((e - s) / 1000)
  if (sec < 60) return `${sec}秒`
  if (sec < 3600) return `${Math.floor(sec / 60)}分${sec % 60}秒`
  return `${Math.floor(sec / 3600)}时${Math.floor((sec % 3600) / 60)}分`
}

function parseErrors(errors: string): string[] {
  try {
    const parsed = JSON.parse(errors)
    if (Array.isArray(parsed)) return parsed
    return []
  } catch {
    return []
  }
}

onMounted(() => {
  loadTasks()
  startPolling()
})

onUnmounted(() => {
  stopPolling()
})
</script>

<style lang="scss" scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', sans-serif;
}

.tasks-card {
  background: white;
  border: 1px solid #E8E8ED;
  border-radius: 12px;
  padding: 24px;
}

.empty-state {
  text-align: center;
  padding: 48px 0;
  color: #86868B;

  p { margin: 12px 0; font-size: 14px; }
  .link { color: #0071E3; font-size: 13px; }
}

.task-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-item {
  border: 1px solid #E8E8ED;
  border-radius: 10px;
  padding: 16px 20px;
  transition: border-color 0.15s;

  &.running { border-left: 3px solid #0071E3; }
  &.completed { border-left: 3px solid #34C759; }
  &.failed { border-left: 3px solid #FF3B30; }
  &.cancelled { border-left: 3px solid #86868B; }
  &.pending { border-left: 3px solid #FF9500; }
}

.task-main {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.task-info { flex: 1; min-width: 0; }

.task-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.task-type { font-size: 14px; font-weight: 500; }

.task-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;

  &.pending { background: #FFF3E0; color: #E65100; }
  &.running { background: #E3F2FD; color: #1565C0; }
  &.completed { background: #E8F5E9; color: #2E7D32; }
  &.failed { background: #FFEBEE; color: #C62828; }
  &.cancelled { background: #F5F5F5; color: #616161; }
}

.task-id { font-size: 12px; color: #AEAEB2; }

.task-source {
  font-size: 13px;
  color: #424245;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 400px;
}

.task-time { font-size: 12px; color: #86868B; margin-top: 2px; }

.task-stats {
  display: flex;
  gap: 16px;
  flex-shrink: 0;
}

.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.stat-val {
  font-size: 20px;
  font-weight: 700;

  &.green { color: #34C759; }
  &.blue { color: #007AFF; }
  &.gray { color: #86868B; }
  &.red { color: #FF3B30; }
}

.stat-lbl { font-size: 11px; color: #86868B; }

.task-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
  align-self: center;
}

.task-progress { margin-top: 12px; }

.task-errors {
  margin-top: 12px;
  padding: 12px;
  background: #FFF8F0;
  border-radius: 6px;
  max-height: 120px;
  overflow-y: auto;
}

.error-item { font-size: 12px; color: #424245; padding: 2px 0; }

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
</style>
