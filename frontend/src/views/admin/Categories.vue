<template>
  <div class="categories-page">
    <div class="page-header">
      <h1 class="page-title">分类管理</h1>
      <div class="header-actions">
        <el-button
          v-if="selectedIds.length > 0"
          type="danger"
          @click="handleBatchDelete"
        >
          <el-icon><Delete /></el-icon>
          批量删除 ({{ selectedIds.length }})
        </el-button>
        <el-button type="primary" @click="openCreateDialog(null)">
          新增分类
        </el-button>
      </div>
    </div>

    <div v-loading="loading">
      <div v-if="flatCategories.length === 0" class="empty-state">
        暂无分类，点击上方按钮创建
      </div>
      <div v-else class="category-table">
        <div class="table-header">
          <el-checkbox
            :model-value="allSelected"
            :indeterminate="someSelected"
            @change="toggleSelectAll"
          />
          <span class="header-label header-drag"></span>
          <span class="header-label header-name">分类名称</span>
          <span class="header-label header-count">文档数</span>
          <span class="header-label header-actions">操作</span>
        </div>
        <div ref="topListRef" class="sortable-list">
          <CategoryRow
            v-for="cat in categories"
            :key="cat.id"
            :category="cat"
            :level="0"
            :selected-ids="selectedIds"
            @edit="openEditDialog"
            @delete="handleDelete"
            @add-child="openCreateDialog"
            @toggle-select="toggleSelect"
            @reorder-children="handleReorderChildren"
          />
        </div>
      </div>
    </div>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="showDialog"
      :title="editingId ? '编辑分类' : '新增分类'"
      width="400px"
    >
      <el-form :model="form" label-position="top">
        <el-form-item label="分类名称">
          <el-input v-model="form.name" placeholder="输入分类名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import Sortable from 'sortablejs'
import { listCategories, createCategory, updateCategory, deleteCategory, batchDeleteCategories, sortCategories } from '@/api/admin'
import type { Category } from '@/types'
import CategoryRow from './components/CategoryRow.vue'

const categories = ref<Category[]>([])
const flatCategories = ref<Category[]>([])
const loading = ref(false)
const selectedIds = ref<number[]>([])
const saving = ref(false)

const topListRef = ref<HTMLElement | null>(null)
let topSortable: Sortable | null = null

const showDialog = ref(false)
const submitLoading = ref(false)
const editingId = ref<number | null>(null)
const parentId = ref<number | null>(null)
const form = reactive({ name: '' })

const allSelected = computed(() =>
  flatCategories.value.length > 0 && selectedIds.value.length === flatCategories.value.length
)
const someSelected = computed(() =>
  selectedIds.value.length > 0 && selectedIds.value.length < flatCategories.value.length
)

// ========== 拖拽排序 ==========

function initTopSortable() {
  if (!topListRef.value) return
  if (topSortable) {
    topSortable.destroy()
  }
  topSortable = Sortable.create(topListRef.value, {
    handle: '.drag-handle',
    animation: 200,
    ghostClass: 'sortable-ghost',
    chosenClass: 'sortable-chosen',
    onEnd(evt) {
      const oldIdx = evt.oldIndex
      const newIdx = evt.newIndex
      if (oldIdx == null || newIdx == null || oldIdx === newIdx) return
      const from: number = oldIdx
      const to: number = newIdx
      const moved = categories.value.splice(from, 1)[0]
      categories.value.splice(to, 0, moved)
      saveOrder()
    },
  })
}

function handleReorderChildren(parentId: number, oldIndex: number, newIndex: number) {
  const parent = findInTree(categories.value, parentId)
  if (!parent?.children || oldIndex === newIndex) return
  const moved = parent.children.splice(oldIndex, 1)[0]
  parent.children.splice(newIndex, 0, moved)
  saveOrder()
}

function findInTree(tree: Category[], id: number): Category | null {
  for (const cat of tree) {
    if (cat.id === id) return cat
    if (cat.children?.length) {
      const found = findInTree(cat.children, id)
      if (found) return found
    }
  }
  return null
}

function collectOrderedIds(tree: Category[]): number[] {
  const ids: number[] = []
  for (const cat of tree) {
    ids.push(cat.id)
    if (cat.children?.length) {
      ids.push(...collectOrderedIds(cat.children))
    }
  }
  return ids
}

async function saveOrder() {
  const ids = collectOrderedIds(categories.value)
  const items = ids.map((id, index) => ({ id, sort_order: index }))
  saving.value = true
  try {
    const res = await sortCategories(items)
    if (res.code === 0) {
      ElMessage.success('排序已保存')
    } else {
      ElMessage.error(res.message || '排序保存失败')
    }
  } catch {
    ElMessage.error('排序保存失败')
  } finally {
    saving.value = false
  }
}

// ========== 数据获取 ==========

async function fetchCategories() {
  loading.value = true
  try {
    const res = await listCategories()
    if (res.code === 0) {
      flatCategories.value = res.data || []
      categories.value = buildTree(res.data || [])
      await nextTick()
      initTopSortable()
    }
  } catch {
    ElMessage.error('获取分类列表失败')
  } finally {
    loading.value = false
  }
}

function buildTree(list: Category[]): Category[] {
  const map = new Map<number, Category>()
  const roots: Category[] = []

  list.forEach((item) => {
    map.set(item.id, { ...item, children: [] })
  })

  list.forEach((item) => {
    const node = map.get(item.id)!
    if (item.parent_id && map.has(item.parent_id)) {
      map.get(item.parent_id)!.children!.push(node)
    } else {
      roots.push(node)
    }
  })

  return roots
}

// ========== 选择 ==========

function toggleSelect(id: number) {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) {
    selectedIds.value.splice(idx, 1)
  } else {
    selectedIds.value.push(id)
  }
}

function toggleSelectAll(val: boolean) {
  if (val) {
    selectedIds.value = flatCategories.value.map(c => c.id)
  } else {
    selectedIds.value = []
  }
}

// ========== CRUD ==========

function openCreateDialog(pId: number | null) {
  editingId.value = null
  parentId.value = pId
  form.name = ''
  showDialog.value = true
}

function openEditDialog(cat: Category) {
  editingId.value = cat.id
  parentId.value = null
  form.name = cat.name
  showDialog.value = true
}

async function handleSubmit() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入分类名称')
    return
  }

  submitLoading.value = true
  try {
    if (editingId.value) {
      const existing = flatCategories.value.find(c => c.id === editingId.value)
      const res = await updateCategory(editingId.value, {
        name: form.name,
        sort_order: existing?.sort_order ?? 0,
        parent_id: existing?.parent_id,
      })
      if (res.code === 0) {
        ElMessage.success('修改成功')
        showDialog.value = false
        fetchCategories()
      } else {
        ElMessage.error(res.message || '修改失败')
      }
    } else {
      const data: any = { name: form.name, sort_order: 9999 }
      if (parentId.value) data.parent_id = parentId.value
      const res = await createCategory(data)
      if (res.code === 0) {
        ElMessage.success('创建成功')
        showDialog.value = false
        fetchCategories()
      } else {
        ElMessage.error(res.message || '创建失败')
      }
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('删除分类将同时删除其下所有子分类，确定？', '确认删除', { type: 'warning' })
    const res = await deleteCategory(id)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      selectedIds.value = selectedIds.value.filter(sid => sid !== id)
      fetchCategories()
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
      `确定删除选中的 ${selectedIds.value.length} 个分类？子分类将同时被删除。`,
      '批量删除',
      { type: 'warning' }
    )
    const res = await batchDeleteCategories(selectedIds.value)
    if (res.code === 0) {
      ElMessage.success(`成功删除 ${res.data?.deleted || 0} 个分类`)
      selectedIds.value = []
      fetchCategories()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch {
    // cancelled
  }
}

onMounted(fetchCategories)

onBeforeUnmount(() => {
  if (topSortable) {
    topSortable.destroy()
    topSortable = null
  }
})
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

.empty-state {
  text-align: center;
  padding: 64px;
  color: #AEAEB2;
  font-size: 14px;
}

.category-table {
  border: 1px solid #F2F2F7;
  border-radius: 8px;
  overflow: hidden;
}

.table-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: #F5F5F7;
  border-bottom: 1px solid #E8E8ED;
  font-size: 13px;
  font-weight: 600;
  color: #86868B;

  .header-label {
    &.header-drag { width: 24px; flex-shrink: 0; }
    &.header-name { flex: 1; }
    &.header-count { width: 60px; text-align: center; }
    &.header-actions { width: 160px; text-align: right; }
  }
}

.sortable-list {
  min-height: 1px;
}

// SortableJS 拖拽样式
:deep(.sortable-ghost) {
  opacity: 0.4;
  background: #F5F5F7;
}

:deep(.sortable-chosen) {
  background: #FAFAFA;
}

:deep(.sortable-drag) {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  border-radius: 8px;
}
</style>
