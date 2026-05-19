<template>
  <div class="credits-page">
    <div class="page-header">
      <h2>致谢管理</h2>
      <el-button type="primary" @click="openDialog()">
        <el-icon><Plus /></el-icon>
        添加致谢
      </el-button>
    </div>

    <el-table :data="credits" v-loading="loading" stripe>
      <el-table-column prop="sort_order" label="排序" width="80" />
      <el-table-column label="图标" width="80">
        <template #default="{ row }">
          <img v-if="row.icon_url" :src="row.icon_url" class="credit-icon" />
          <span v-else class="no-icon">—</span>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="名称" min-width="150" />
      <el-table-column prop="url" label="URL" min-width="200">
        <template #default="{ row }">
          <a :href="row.url" target="_blank" class="url-link">{{ row.url }}</a>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
      <el-table-column prop="license" label="许可证" width="120" />
      <el-table-column prop="stars" label="Stars" width="100" />
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="editingId ? '编辑致谢' : '添加致谢'"
      width="560px"
      destroy-on-close
    >
      <el-form :model="form" label-width="90px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="项目名称" />
        </el-form-item>
        <el-form-item label="URL" required>
          <el-input v-model="form.url" placeholder="https://github.com/..." />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="简短描述" />
        </el-form-item>
        <el-form-item label="图标 URL">
          <el-input v-model="form.icon_url" placeholder="https://..." />
        </el-form-item>
        <el-form-item label="许可证">
          <el-input v-model="form.license" placeholder="MIT / Apache-2.0" />
        </el-form-item>
        <el-form-item label="Stars">
          <el-input v-model="form.stars" placeholder="1.2k" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort_order" :min="0" :max="9999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAdminCredits, createCredit, updateCredit, deleteCredit } from '@/api/admin'
import type { Credit } from '@/types'

const credits = ref<Credit[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)

const defaultForm = {
  name: '',
  url: '',
  description: '',
  icon_url: '',
  license: '',
  stars: '',
  sort_order: 0,
}

const form = ref({ ...defaultForm })

async function fetchCredits() {
  loading.value = true
  try {
    const res = await getAdminCredits()
    if (res.code === 0) {
      credits.value = res.data || []
    }
  } catch (e) {
    ElMessage.error('获取致谢列表失败')
  } finally {
    loading.value = false
  }
}

function openDialog(row?: Credit) {
  if (row) {
    editingId.value = row.id
    form.value = {
      name: row.name,
      url: row.url,
      description: row.description,
      icon_url: row.icon_url,
      license: row.license,
      stars: row.stars,
      sort_order: row.sort_order,
    }
  } else {
    editingId.value = null
    form.value = { ...defaultForm }
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.value.name || !form.value.url) {
    ElMessage.warning('名称和 URL 为必填项')
    return
  }
  saving.value = true
  try {
    if (editingId.value) {
      const res = await updateCredit(editingId.value, form.value)
      if (res.code === 0) {
        ElMessage.success('更新成功')
        dialogVisible.value = false
        fetchCredits()
      } else {
        ElMessage.error(res.message || '更新失败')
      }
    } else {
      const res = await createCredit(form.value)
      if (res.code === 0) {
        ElMessage.success('创建成功')
        dialogVisible.value = false
        fetchCredits()
      } else {
        ElMessage.error(res.message || '创建失败')
      }
    }
  } catch (e) {
    ElMessage.error('操作失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: Credit) {
  try {
    await ElMessageBox.confirm(`确定删除「${row.name}」？`, '确认删除', {
      type: 'warning',
    })
    const res = await deleteCredit(row.id)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      fetchCredits()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch {
    // 取消
  }
}

onMounted(fetchCredits)
</script>

<style lang="scss" scoped>
.credits-page {
  max-width: 1200px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;

  h2 {
    font-size: 24px;
    font-weight: 600;
    color: #1D1D1F;
    margin: 0;
  }
}

.credit-icon {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  object-fit: contain;
}

.no-icon {
  color: #AEAEB2;
}

.url-link {
  color: #0071E3;
  text-decoration: none;
  font-size: 13px;
  word-break: break-all;

  &:hover {
    text-decoration: underline;
  }
}
</style>
