<template>
  <div class="users-page">
    <div class="page-header">
      <h1 class="page-title">用户管理</h1>
    </div>

    <el-table :data="users" style="width: 100%" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="用户名" />
      <el-table-column label="角色" width="150">
        <template #default="{ row }">
          <el-select
            :model-value="row.role"
            size="small"
            @change="(val: string) => handleRoleChange(row.id, val)"
          >
            <el-option label="管理员" value="admin" />
            <el-option label="编辑者" value="editor" />
            <el-option label="读者" value="reader" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
            {{ row.status === 'active' ? '正常' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="注册时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button
            type="primary"
            link
            size="small"
            @click="handleStatusChange(row.id, row.status === 'active' ? 'disabled' : 'active')"
          >
            {{ row.status === 'active' ? '禁用' : '启用' }}
          </el-button>
          <el-button
            type="danger"
            link
            size="small"
            @click="handleDelete(row.id, row.username)"
          >
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
        @current-change="fetchUsers"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listUsers, updateUserRole, updateUserStatus, deleteUser } from '@/api/admin'
import type { User } from '@/types'

const users = ref<User[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = 20
const total = ref(0)

async function fetchUsers() {
  loading.value = true
  try {
    const res = await listUsers(page.value, pageSize)
    if (res.code === 0) {
      users.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } catch {
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

async function handleRoleChange(id: number, role: string) {
  try {
    await ElMessageBox.confirm(`确定将用户角色修改为 ${role}？`, '确认修改', { type: 'warning' })
    const res = await updateUserRole(id, role)
    if (res.code === 0) {
      ElMessage.success('角色修改成功')
      fetchUsers()
    } else {
      ElMessage.error(res.message || '修改失败')
    }
  } catch {
    // cancelled
  }
}

async function handleStatusChange(id: number, status: string) {
  const label = status === 'active' ? '启用' : '禁用'
  try {
    await ElMessageBox.confirm(`确定${label}该用户？`, '确认操作', { type: 'warning' })
    const res = await updateUserStatus(id, status)
    if (res.code === 0) {
      ElMessage.success(`${label}成功`)
      fetchUsers()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch {
    // cancelled
  }
}

async function handleDelete(id: number, username: string) {
  try {
    await ElMessageBox.confirm(`确定删除用户「${username}」？此操作不可恢复！`, '确认删除', { type: 'error', confirmButtonText: '确认删除', cancelButtonText: '取消' })
    const res = await deleteUser(id)
    if (res.code === 0) {
      ElMessage.success('用户已删除')
      fetchUsers()
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

onMounted(fetchUsers)
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

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 24px;
}
</style>
