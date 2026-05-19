<template>
  <div class="documents-page">
    <div class="page-header">
      <h1 class="page-title">文档管理</h1>
      <div class="header-actions">
        <el-button
          v-if="selectedIds.length > 0"
          type="danger"
          @click="handleBatchDelete"
        >
          <el-icon><Delete /></el-icon>
          批量删除 ({{ selectedIds.length }})
        </el-button>
        <el-button type="primary" @click="handleCreate">
          新建文档
        </el-button>
      </div>
    </div>

    <el-table
      :data="documents"
      style="width: 100%"
      v-loading="loading"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="50" />
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="title" label="标题" min-width="200" />
      <el-table-column label="分类" width="150">
        <template #default="{ row }">
          {{ row.category_name || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="作者" width="120">
        <template #default="{ row }">
          {{ row.author_name || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.is_published ? 'success' : 'info'" size="small">
            {{ row.is_published ? '已发布' : '草稿' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="更新时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.updated_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220">
        <template #default="{ row }">
          <el-button type="primary" link size="small" @click="handleEdit(row.id)">
            编辑
          </el-button>
          <el-button
            type="primary"
            link
            size="small"
            @click="handlePublish(row.id, !row.is_published)"
          >
            {{ row.is_published ? '取消发布' : '发布' }}
          </el-button>
          <el-button type="danger" link size="small" @click="handleDelete(row.id)">
            删除
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
        @current-change="fetchDocuments"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listDocuments, deleteDocument, batchDeleteDocuments, publishDocument, createDocument } from '@/api/admin'
import type { Document } from '@/types'

const router = useRouter()
const documents = ref<Document[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = 20
const total = ref(0)
const selectedIds = ref<number[]>([])

async function fetchDocuments() {
  loading.value = true
  try {
    const res = await listDocuments({ page: page.value, size: pageSize })
    if (res.code === 0) {
      documents.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } catch {
    ElMessage.error('获取文档列表失败')
  } finally {
    loading.value = false
  }
}

function handleSelectionChange(rows: Document[]) {
  selectedIds.value = rows.map(r => r.id)
}

async function handleCreate() {
  try {
    const res = await createDocument({ title: '无标题文档', content: '' })
    if (res.code === 0 && res.data) {
      router.push(`/admin/dashboard/documents/${res.data.id}`)
    } else {
      ElMessage.error(res.message || '创建失败')
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '创建失败')
  }
}

function handleEdit(id: number) {
  router.push(`/admin/dashboard/documents/${id}`)
}

async function handlePublish(id: number, is_published: boolean) {
  const label = is_published ? '发布' : '取消发布'
  try {
    await ElMessageBox.confirm(`确定${label}该文档？`, '确认操作', { type: 'warning' })
    const res = await publishDocument(id, is_published)
    if (res.code === 0) {
      ElMessage.success(`${label}成功`)
      fetchDocuments()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch {
    // cancelled
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定删除该文档？此操作不可恢复。', '确认删除', { type: 'warning' })
    const res = await deleteDocument(id)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      fetchDocuments()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch {
    // cancelled
  }
}

async function handleBatchDelete() {
  try {
    await ElMessageBox.confirm(
      `确定删除选中的 ${selectedIds.value.length} 篇文档？此操作不可恢复。`,
      '批量删除',
      { type: 'warning' }
    )
    const res = await batchDeleteDocuments(selectedIds.value)
    if (res.code === 0) {
      ElMessage.success(`成功删除 ${res.data?.deleted || 0} 篇文档`)
      selectedIds.value = []
      fetchDocuments()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch {
    // cancelled
  }
}

function formatTime(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(fetchDocuments)
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
  gap: 12px;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 24px;
}
</style>
