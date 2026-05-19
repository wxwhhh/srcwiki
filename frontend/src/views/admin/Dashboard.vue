<template>
  <div class="dashboard-page">
    <h1 class="page-title">概览</h1>

    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-number">{{ stats.user_count }}</div>
        <div class="stat-label">用户数</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">{{ stats.doc_count }}</div>
        <div class="stat-label">文档数</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">{{ stats.category_count }}</div>
        <div class="stat-label">分类数</div>
      </div>
    </div>

    <div class="recent-section">
      <h2 class="section-title">最近操作</h2>
      <div v-if="loading" class="loading-state">
        <el-icon class="is-loading"><Loading /></el-icon>
      </div>
      <div v-else-if="stats.recent_logs.length === 0" class="empty-state">
        暂无操作记录
      </div>
      <el-table v-else :data="stats.recent_logs" style="width: 100%">
        <el-table-column prop="username" label="用户" width="120" />
        <el-table-column prop="action" label="操作" width="150" />
        <el-table-column prop="target_type" label="目标" width="120" />
        <el-table-column prop="ip" label="IP" width="140" />
        <el-table-column label="时间">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { getDashboardStats } from '@/api/admin'
import type { AuditLog } from '@/types'

const loading = ref(true)
const stats = reactive({
  user_count: 0,
  doc_count: 0,
  category_count: 0,
  recent_logs: [] as AuditLog[],
})

onMounted(async () => {
  try {
    const res = await getDashboardStats()
    if (res.code === 0) {
      Object.assign(stats, res.data)
    }
  } catch {
    // ignore
  } finally {
    loading.value = false
  }
})

function formatTime(dateStr: string): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleString('zh-CN')
}
</script>

<style lang="scss" scoped>
.dashboard-page {}

.page-title {
  font-size: 36px;
  font-weight: 600;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', sans-serif;
  margin-bottom: 32px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
  margin-bottom: 48px;
}

.stat-card {
  background: white;
  padding: 32px;
  border-radius: 12px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

.stat-number {
  font-size: 48px;
  font-weight: 700;
  color: #1D1D1F;
  line-height: 1;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: #86868B;
}

.section-title {
  font-size: 22px;
  font-weight: 600;
  margin-bottom: 16px;
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 32px;
  color: #AEAEB2;
}

.empty-state {
  text-align: center;
  padding: 32px;
  color: #AEAEB2;
  font-size: 14px;
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
