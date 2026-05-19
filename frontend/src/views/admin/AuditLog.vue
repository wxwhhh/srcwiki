<template>
  <div class="audit-page">
    <h1 class="page-title">审计日志</h1>

    <!-- Tab 切换 -->
    <div class="tab-bar">
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'stats' }"
        @click="activeTab = 'stats'"
      >
        登录统计
      </button>
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'logs' }"
        @click="activeTab = 'logs'"
      >
        操作日志
      </button>
    </div>

    <!-- 登录统计 Tab -->
    <div v-if="activeTab === 'stats'" class="tab-content">
      <!-- 统计卡片 -->
      <div class="stats-cards">
        <div class="stat-card">
          <div class="stat-value">{{ stats?.total_logins ?? '-' }}</div>
          <div class="stat-label">总登录次数</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ stats?.today_logins ?? '-' }}</div>
          <div class="stat-label">今日登录</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ stats?.active_users ?? '-' }}</div>
          <div class="stat-label">活跃用户</div>
        </div>
      </div>

      <!-- 用户登录排行 -->
      <h2 class="section-title">用户登录排行</h2>
      <el-table :data="stats?.logins_by_user ?? []" style="width: 100%" v-loading="statsLoading">
        <el-table-column prop="username" label="用户名" width="160" />
        <el-table-column prop="count" label="登录次数" width="120" sortable />
        <el-table-column prop="last_ip" label="最后登录 IP" width="160" />
        <el-table-column label="最后登录时间">
          <template #default="{ row }">
            {{ formatTime(row.last_login) }}
          </template>
        </el-table-column>
      </el-table>

      <!-- 最近登录记录 -->
      <h2 class="section-title">最近登录记录</h2>
      <el-table :data="stats?.recent_logins ?? []" style="width: 100%" v-loading="statsLoading">
        <el-table-column prop="username" label="用户名" width="140" />
        <el-table-column prop="ip" label="IP 地址" width="160" />
        <el-table-column prop="user_agent" label="浏览器" min-width="200" show-overflow-tooltip />
        <el-table-column label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 操作日志 Tab -->
    <div v-if="activeTab === 'logs'" class="tab-content">
      <div class="filter-bar">
        <el-input
          v-model="filters.action"
          placeholder="操作类型"
          clearable
          class="filter-input"
        />
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          class="filter-date"
          @change="handleDateChange"
        />
        <el-button type="primary" @click="fetchLogs">查询</el-button>
      </div>

      <el-table :data="logs" style="width: 100%" v-loading="logsLoading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户" width="120" />
        <el-table-column prop="action" label="操作" width="150" />
        <el-table-column prop="target_type" label="目标类型" width="120" />
        <el-table-column prop="target_id" label="目标ID" width="100" />
        <el-table-column prop="detail" label="详情" min-width="200" show-overflow-tooltip />
        <el-table-column prop="ip" label="IP" width="140" />
        <el-table-column label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="fetchLogs"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { listAuditLogs, getLoginStats } from '@/api/admin'
import type { LoginStats } from '@/api/admin'
import type { AuditLog } from '@/types'

// ─── Tab ───
const activeTab = ref<'stats' | 'logs'>('stats')

// ─── 登录统计 ───
const stats = ref<LoginStats | null>(null)
const statsLoading = ref(false)

async function fetchStats() {
  statsLoading.value = true
  try {
    const res = await getLoginStats()
    if (res.code === 0) {
      stats.value = res.data
    }
  } catch {
    ElMessage.error('获取登录统计失败')
  } finally {
    statsLoading.value = false
  }
}

// ─── 操作日志 ───
const logs = ref<AuditLog[]>([])
const logsLoading = ref(false)
const page = ref(1)
const pageSize = 20
const total = ref(0)
const dateRange = ref<[Date, Date] | null>(null)

const filters = reactive({
  action: '',
  start_date: '',
  end_date: '',
})

async function fetchLogs() {
  logsLoading.value = true
  try {
    const params: any = {
      page: page.value,
      page_size: pageSize,
    }
    if (filters.action) params.action = filters.action
    if (filters.start_date) params.start_date = filters.start_date
    if (filters.end_date) params.end_date = filters.end_date

    const res = await listAuditLogs(params)
    if (res.code === 0) {
      logs.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } catch {
    ElMessage.error('获取审计日志失败')
  } finally {
    logsLoading.value = false
  }
}

function handleDateChange(dates: [Date, Date] | null) {
  if (dates) {
    filters.start_date = dates[0].toISOString().split('T')[0]
    filters.end_date = dates[1].toISOString().split('T')[0]
  } else {
    filters.start_date = ''
    filters.end_date = ''
  }
}

function formatTime(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

// 切换 tab 时加载对应数据
watch(activeTab, (tab) => {
  if (tab === 'stats' && !stats.value) fetchStats()
  if (tab === 'logs' && logs.value.length === 0) fetchLogs()
})

onMounted(fetchStats)
</script>

<style lang="scss" scoped>
.page-title {
  font-size: 28px;
  font-weight: 600;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', sans-serif;
  margin-bottom: 20px;
}

/* ─── Tab Bar ─── */
.tab-bar {
  display: flex;
  gap: 4px;
  border-bottom: 1px solid #E8E8ED;
  margin-bottom: 24px;
}

.tab-btn {
  padding: 10px 20px;
  font-size: 14px;
  font-weight: 500;
  color: #86868B;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;

  &:hover {
    color: #1D1D1F;
  }

  &.active {
    color: #0071E3;
    border-bottom-color: #0071E3;
  }
}

.tab-content {
  animation: fadeIn 0.15s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(4px); }
  to   { opacity: 1; transform: translateY(0); }
}

/* ─── 统计卡片 ─── */
.stats-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 32px;
}

.stat-card {
  background: #F2F2F7;
  border-radius: 12px;
  padding: 24px;
  text-align: center;
}

.stat-value {
  font-size: 36px;
  font-weight: 700;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', sans-serif;
  color: #1D1D1F;
  line-height: 1.2;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 13px;
  color: #86868B;
}

/* ─── Section ─── */
.section-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 12px;
  margin-top: 8px;
}

/* ─── 操作日志 ─── */
.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.filter-input {
  width: 200px;
}

.filter-date {
  width: 280px;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 24px;
}

/* ─── 响应式 ─── */
@media (max-width: 700px) {
  .stats-cards {
    grid-template-columns: 1fr;
  }
}
</style>
