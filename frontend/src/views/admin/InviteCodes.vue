<template>
  <div class="invites-page">
    <div class="page-header">
      <h1 class="page-title">邀请码管理</h1>
      <div class="header-actions">
        <el-button type="primary" @click="showCreateDialog = true">
          生成邀请码
        </el-button>
        <el-button type="success" @click="showBatchDialog = true">
          批量生成
        </el-button>
      </div>
    </div>

    <el-table :data="invites" style="width: 100%" v-loading="loading">
      <el-table-column prop="code" label="邀请码" width="200">
        <template #default="{ row }">
          <code class="invite-code">{{ row.code }}</code>
        </template>
      </el-table-column>
      <el-table-column label="角色" width="100">
        <template #default="{ row }">
          <el-tag size="small">{{ row.role === 'editor' ? '编辑者' : '读者' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="使用情况" width="120">
        <template #default="{ row }">
          {{ row.use_count }} / {{ row.max_uses }}
        </template>
      </el-table-column>
      <el-table-column label="过期时间" width="180">
        <template #default="{ row }">
          {{ row.expires_at ? formatTime(row.expires_at) : '永不过期' }}
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="isExpired(row) ? 'danger' : 'success'" size="small">
            {{ isExpired(row) ? '已过期' : '有效' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button type="primary" link size="small" @click="handleDelete(row.id)">
            作废
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="fetchInvites"
      />
    </div>

    <!-- 单个创建对话框 -->
    <el-dialog v-model="showCreateDialog" title="生成邀请码" width="400px">
      <el-form :model="createForm" label-position="top">
        <el-form-item label="角色">
          <el-select v-model="createForm.role" style="width: 100%">
            <el-option label="编辑者" value="editor" />
            <el-option label="读者" value="reader" />
          </el-select>
        </el-form-item>
        <el-form-item label="最大使用次数">
          <el-input-number v-model="createForm.max_uses" :min="1" :max="100" />
        </el-form-item>
        <el-form-item label="过期时间（可选）">
          <el-date-picker
            v-model="createForm.expires_at"
            type="datetime"
            placeholder="选择过期时间"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="handleCreate">
          生成
        </el-button>
      </template>
    </el-dialog>

    <!-- 批量生成对话框 -->
    <el-dialog v-model="showBatchDialog" title="批量生成邀请码" width="450px">
      <el-form :model="batchForm" label-position="top">
        <el-form-item label="生成数量">
          <el-input-number v-model="batchForm.count" :min="1" :max="500" style="width: 100%" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="batchForm.role" style="width: 100%">
            <el-option label="编辑者" value="editor" />
            <el-option label="读者" value="reader" />
          </el-select>
        </el-form-item>
        <el-form-item label="有效期（小时）">
          <el-input-number v-model="batchForm.expires_in_hours" :min="0" style="width: 100%" />
          <div class="form-tip">0 表示永不过期</div>
        </el-form-item>
        <el-form-item label="最大使用次数">
          <el-input-number v-model="batchForm.max_uses" :min="0" :max="100" style="width: 100%" />
          <div class="form-tip">0 表示不限（默认 1 次）</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBatchDialog = false">取消</el-button>
        <el-button type="primary" :loading="batchLoading" @click="handleBatchCreate">
          批量生成
        </el-button>
      </template>
    </el-dialog>

    <!-- 批量生成结果对话框 -->
    <el-dialog v-model="showResultDialog" title="批量生成结果" width="650px">
      <div class="result-header">
        <span>成功生成 <strong>{{ batchResults.length }}</strong> 个邀请码</span>
        <div class="result-actions">
          <el-button size="small" @click="copyAllCodes">复制全部</el-button>
          <el-button size="small" @click="exportCodes">导出 CSV</el-button>
        </div>
      </div>
      <el-table :data="batchResults" style="width: 100%" max-height="400">
        <el-table-column prop="code" label="邀请码" width="200">
          <template #default="{ row }">
            <code class="invite-code">{{ row.code }}</code>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.role === 'editor' ? '编辑者' : '读者' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="max_uses" label="最大使用" width="100" />
        <el-table-column label="过期时间">
          <template #default="{ row }">
            {{ row.expires_at ? formatTime(row.expires_at) : '永不过期' }}
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button type="primary" @click="showResultDialog = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listInvites, createInvite, batchCreateInvites, deleteInvite } from '@/api/admin'
import type { InviteCode } from '@/types'

const invites = ref<InviteCode[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = 20
const total = ref(0)

// 单个创建
const showCreateDialog = ref(false)
const createLoading = ref(false)
const createForm = reactive({
  role: 'reader',
  max_uses: 1,
  expires_at: null as Date | null,
})

// 批量创建
const showBatchDialog = ref(false)
const batchLoading = ref(false)
const batchForm = reactive({
  count: 10,
  role: 'reader',
  expires_in_hours: 0,
  max_uses: 1,
})

// 批量结果
const showResultDialog = ref(false)
const batchResults = ref<InviteCode[]>([])

async function fetchInvites() {
  loading.value = true
  try {
    const res = await listInvites(page.value, pageSize)
    if (res.code === 0) {
      invites.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } catch {
    ElMessage.error('获取邀请码列表失败')
  } finally {
    loading.value = false
  }
}

async function handleCreate() {
  createLoading.value = true
  try {
    const data: any = {
      role: createForm.role,
      max_uses: createForm.max_uses,
    }
    if (createForm.expires_at) {
      data.expires_at = createForm.expires_at.toISOString()
    }
    const res = await createInvite(data)
    if (res.code === 0) {
      ElMessage.success('邀请码生成成功')
      showCreateDialog.value = false
      createForm.role = 'reader'
      createForm.max_uses = 1
      createForm.expires_at = null
      fetchInvites()
    } else {
      ElMessage.error(res.message || '生成失败')
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '生成失败')
  } finally {
    createLoading.value = false
  }
}

async function handleBatchCreate() {
  batchLoading.value = true
  try {
    const res = await batchCreateInvites({
      count: batchForm.count,
      role: batchForm.role,
      expires_in_hours: batchForm.expires_in_hours,
      max_uses: batchForm.max_uses,
    })
    if (res.code === 0) {
      ElMessage.success(`成功生成 ${res.data.count} 个邀请码`)
      showBatchDialog.value = false
      batchResults.value = res.data.codes || []
      showResultDialog.value = true
      fetchInvites()
    } else {
      ElMessage.error(res.message || '批量生成失败')
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '批量生成失败')
  } finally {
    batchLoading.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定作废该邀请码？', '确认操作', { type: 'warning' })
    const res = await deleteInvite(id)
    if (res.code === 0) {
      ElMessage.success('已作废')
      fetchInvites()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch {
    // cancelled
  }
}

function copyAllCodes() {
  const text = batchResults.value.map(c => c.code).join('\n')
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('已复制全部邀请码')
  }).catch(() => {
    ElMessage.error('复制失败，请手动复制')
  })
}

function exportCodes() {
  const header = '邀请码,角色,最大使用次数,过期时间\n'
  const rows = batchResults.value.map(c =>
    `${c.code},${c.role === 'editor' ? '编辑者' : '读者'},${c.max_uses},${c.expires_at || '永不过期'}`
  ).join('\n')
  const csv = '\uFEFF' + header + rows
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `invite_codes_${new Date().toISOString().slice(0, 10)}.csv`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('已导出 CSV 文件')
}

function isExpired(invite: InviteCode): boolean {
  if (!invite.expires_at) return false
  return new Date(invite.expires_at) < new Date()
}

function formatTime(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(fetchInvites)
</script>

<style lang="scss" scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', sans-serif;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.invite-code {
  font-family: 'SF Mono', monospace;
  font-size: 13px;
  background: #F5F5F7;
  padding: 2px 8px;
  border-radius: 4px;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 24px;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.result-actions {
  display: flex;
  gap: 8px;
}
</style>
