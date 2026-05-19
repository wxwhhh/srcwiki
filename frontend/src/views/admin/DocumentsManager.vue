<template>
  <div class="documents-manager">
    <!-- 顶栏 -->
    <div class="page-header">
      <h1 class="page-title">
        <el-button
          v-if="editingDoc"
          text
          class="back-btn"
          @click="backToList"
        >
          <el-icon><Back /></el-icon>
        </el-button>
        {{ editingDoc ? '编辑文档' : '文档管理' }}
      </h1>
      <div class="header-actions">
        <el-button @click="openCreateCategory()">
          <el-icon><FolderAdd /></el-icon>
          新建分类
        </el-button>
        <el-button type="primary" @click="handleCreateDoc">
          <el-icon><DocumentAdd /></el-icon>
          新建文档
        </el-button>
      </div>
    </div>

    <div class="manager-body">
      <!-- 左侧：分类树 -->
      <aside class="category-panel">
        <div class="panel-header">
          <span class="panel-title">分类</span>
          <el-input
            v-model="searchText"
            placeholder="搜索分类..."
            size="small"
            clearable
            class="search-input"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>

        <div ref="treeListRef" class="tree-container" v-loading="categoriesLoading">
          <!-- 全部文档 -->
          <div
            class="tree-node root-node"
            :class="{ active: selectedCategoryId === null && !editingDoc }"
            @click="selectCategory(null)"
          >
            <el-icon :size="14"><Files /></el-icon>
            <span class="node-name">全部文档</span>
            <span class="node-count">{{ totalDocCount }}</span>
          </div>

          <!-- 分类树 -->
          <CategoryTreeNode
            v-for="cat in filteredCategories"
            :key="cat.id"
            :category="cat"
            :level="0"
            :selected-id="selectedCategoryId"
            :search-text="searchText"
            :checked-ids="checkedCategoryIds"
            @select="selectCategory"
            @edit="openEditCategory"
            @delete="handleDeleteCategory"
            @add-child="openCreateCategory"
            @toggle-check="toggleCategoryCheck"
          />

          <div v-if="categories.length === 0 && !categoriesLoading" class="empty-tree">
            <el-icon :size="32" color="#C0C4CC"><FolderOpened /></el-icon>
            <p>暂无分类</p>
          </div>
        </div>

        <!-- 批量删除按钮 -->
        <transition name="fade">
          <div v-if="checkedCategoryIds.size > 0" class="batch-delete-bar">
            <span class="checked-count">已选 {{ checkedCategoryIds.size }} 个目录</span>
            <el-button
              type="danger"
              size="small"
              @click="handleCascadeDelete"
            >
              <el-icon><Delete /></el-icon>
              删除选中
            </el-button>
          </div>
        </transition>
      </aside>

      <!-- 右侧：文档列表 / 文档编辑 -->
      <main class="content-panel">
        <!-- 文档编辑模式 -->
        <div v-if="editingDoc" class="editor-view">
          <div class="edit-meta">
            <el-input
              v-model="editForm.title"
              placeholder="文档标题"
              class="title-input"
              size="large"
            />
            <el-select
              v-model="editForm.category_id"
              placeholder="选择分类"
              clearable
              class="category-select"
            >
              <el-option
                v-for="cat in flatCategoryOptions"
                :key="cat.id"
                :label="cat.label"
                :value="cat.id"
              />
            </el-select>
          </div>

          <div class="editor-body">
            <div class="editor-pane">
              <div class="pane-header">编辑</div>
              <textarea
                v-model="editForm.content"
                class="editor-textarea"
                placeholder="输入 Markdown 内容..."
              />
              <div class="upload-area">
                <el-upload
                  :show-file-list="false"
                  :http-request="handleUpload"
                  accept="image/*"
                >
                  <el-button size="small" text>
                    <el-icon><Upload /></el-icon>
                    上传图片
                  </el-button>
                </el-upload>
              </div>
            </div>
            <div class="preview-pane">
              <div class="pane-header">预览</div>
              <div class="preview-content">
                <MarkdownViewer :content="editForm.content" />
              </div>
            </div>
          </div>

          <div class="edit-footer">
            <div class="footer-left">
              <span class="word-count">字数：{{ wordCount }}</span>
              <span class="save-status" :class="{ unsaved: saveStatus === '未保存' }">
                {{ saveStatus }}
              </span>
            </div>
            <div class="footer-right">
              <el-button @click="handleSave(false)">保存草稿</el-button>
              <el-button type="primary" :loading="saving" @click="handleSave(true)">
                {{ editForm.is_published ? '更新并发布' : '发布' }}
              </el-button>
            </div>
          </div>
        </div>

        <!-- 文档列表模式 -->
        <div v-else class="list-view">
          <div class="list-toolbar">
            <div class="toolbar-left">
              <span class="doc-count-label">
                共 {{ total }} 篇文档
                <template v-if="selectedCategoryId">
                  — {{ selectedCategoryName }}
                </template>
              </span>
            </div>
            <div class="toolbar-right">
              <el-input
                v-model="docSearch"
                placeholder="搜索文档..."
                size="small"
                clearable
                class="doc-search"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
              </el-input>
              <el-select
                v-model="statusFilter"
                placeholder="状态"
                size="small"
                clearable
                class="status-filter"
              >
                <el-option label="已发布" value="published" />
                <el-option label="草稿" value="draft" />
              </el-select>
              <el-button
                v-if="selectedDocIds.length > 0"
                type="danger"
                size="small"
                @click="handleBatchDelete"
              >
                批量删除 ({{ selectedDocIds.length }})
              </el-button>
            </div>
          </div>

          <el-table
            :data="documents"
            style="width: 100%"
            v-loading="docsLoading"
            @selection-change="handleSelectionChange"
            @row-click="handleRowClick"
            class="doc-table"
            row-class-name="doc-row"
          >
            <el-table-column type="selection" width="45" @click.stop />
            <el-table-column prop="title" label="标题" min-width="200">
              <template #default="{ row }">
                <span class="doc-title">{{ row.title }}</span>
              </template>
            </el-table-column>
            <el-table-column label="分类" width="150">
              <template #default="{ row }">
                <span class="doc-category">{{ row.category_name || '未分类' }}</span>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="90" align="center">
              <template #default="{ row }">
                <el-tag
                  :type="row.is_published ? 'success' : 'info'"
                  size="small"
                  effect="light"
                >
                  {{ row.is_published ? '已发布' : '草稿' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="更新时间" width="170">
              <template #default="{ row }">
                <span class="doc-time">{{ formatTime(row.updated_at) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click.stop="openEditor(row)">
                  编辑
                </el-button>
                <el-button
                  type="primary"
                  link
                  size="small"
                  @click.stop="handlePublish(row)"
                >
                  {{ row.is_published ? '取消发布' : '发布' }}
                </el-button>
                <el-button
                  type="danger"
                  link
                  size="small"
                  @click.stop="handleDeleteDoc(row.id)"
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
              @current-change="fetchDocuments"
            />
          </div>
        </div>
      </main>
    </div>

    <!-- 创建/编辑分类对话框 -->
    <el-dialog
      v-model="showCategoryDialog"
      :title="editingCategoryId ? '编辑分类' : '新建分类'"
      width="400px"
    >
      <el-form :model="categoryForm" label-position="top">
        <el-form-item label="分类名称">
          <el-input v-model="categoryForm.name" placeholder="输入分类名称" />
        </el-form-item>
        <el-form-item v-if="categoryForm.parent_id" label="父分类">
          <el-tag>{{ parentCategoryName }}</el-tag>
        </el-form-item>

      </el-form>
      <template #footer>
        <el-button @click="showCategoryDialog = false">取消</el-button>
        <el-button type="primary" :loading="categorySubmitting" @click="handleCategorySubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import Sortable from 'sortablejs'
import {
  listCategories,
  createCategory,
  updateCategory,
  deleteCategory,
  cascadeDeleteCategories,
  sortCategories,
  listDocuments,
  getAdminDocument,
  createDocument,
  updateDocument,
  deleteDocument,
  batchDeleteDocuments,
  publishDocument,
  uploadFile,
} from '@/api/admin'
import type { Category, Document } from '@/types'
import MarkdownViewer from '@/components/MarkdownViewer.vue'
import CategoryTreeNode from './components/CategoryTreeNode.vue'

// ==================== State ====================

// 分类
const categories = ref<Category[]>([])
const flatCategories = ref<Category[]>([])
const categoriesLoading = ref(false)
const selectedCategoryId = ref<number | null>(null)
const searchText = ref('')
const checkedCategoryIds = ref<Set<number>>(new Set())
const treeListRef = ref<HTMLElement | null>(null)
let topSortable: Sortable | null = null
const sortSaving = ref(false)

// 文档列表
const documents = ref<Document[]>([])
const docsLoading = ref(false)
const page = ref(1)
const pageSize = 20
const total = ref(0)
const selectedDocIds = ref<number[]>([])
const docSearch = ref('')
const statusFilter = ref('')

// 文档编辑
const editingDoc = ref(false)
const editingDocId = ref<number | null>(null)
const saving = ref(false)
const saveStatus = ref('')
const editForm = reactive({
  title: '',
  content: '',
  category_id: null as number | null,
  is_published: false,
})

// 分类对话框
const showCategoryDialog = ref(false)
const editingCategoryId = ref<number | null>(null)
const categorySubmitting = ref(false)
const categoryForm = reactive({
  name: '',
  parent_id: null as number | null,
})

// ==================== Computed ====================

const totalDocCount = computed(() => {
  return flatCategories.value.reduce((sum, c) => sum + (c.doc_count || 0), 0)
})

const filteredCategories = computed(() => {
  if (!searchText.value) return categories.value
  const keyword = searchText.value.toLowerCase()
  return filterTree(categories.value, keyword)
})

function filterTree(cats: Category[], keyword: string): Category[] {
  const result: Category[] = []
  for (const cat of cats) {
    const children = cat.children ? filterTree(cat.children, keyword) : []
    if (cat.name.toLowerCase().includes(keyword) || children.length > 0) {
      result.push({ ...cat, children })
    }
  }
  return result
}

const selectedCategoryName = computed(() => {
  if (selectedCategoryId.value === null) return ''
  const cat = flatCategories.value.find(c => c.id === selectedCategoryId.value)
  return cat ? cat.name : ''
})

const parentCategoryName = computed(() => {
  if (!categoryForm.parent_id) return ''
  const cat = flatCategories.value.find(c => c.id === categoryForm.parent_id)
  return cat ? cat.name : ''
})

const flatCategoryOptions = computed(() => {
  return flattenForSelect(categories.value, '')
})

function flattenForSelect(cats: Category[], prefix: string): { id: number; label: string }[] {
  const result: { id: number; label: string }[] = []
  for (const cat of cats) {
    const label = prefix ? `${prefix} / ${cat.name}` : cat.name
    result.push({ id: cat.id, label })
    if (cat.children && cat.children.length > 0) {
      result.push(...flattenForSelect(cat.children, label))
    }
  }
  return result
}

const wordCount = computed(() => {
  return editForm.content.replace(/\s/g, '').length
})

// ==================== 分类操作 ====================

async function fetchCategories() {
  categoriesLoading.value = true
  try {
    const res = await listCategories()
    if (res.code === 0) {
      flatCategories.value = res.data || []
      categories.value = buildTree(res.data || [])
      await nextTick()
      initTopSortable()
    }
  } catch {
    ElMessage.error('获取分类失败')
  } finally {
    categoriesLoading.value = false
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

function selectCategory(id: number | null) {
  selectedCategoryId.value = id
  page.value = 1
  fetchDocuments()
}

function openCreateCategory(parentId: number | null = null) {
  editingCategoryId.value = null
  categoryForm.name = ''
  categoryForm.parent_id = parentId
  showCategoryDialog.value = true
}

function openEditCategory(cat: Category) {
  editingCategoryId.value = cat.id
  categoryForm.name = cat.name
  categoryForm.parent_id = null
  showCategoryDialog.value = true
}

async function handleCategorySubmit() {
  if (!categoryForm.name.trim()) {
    ElMessage.warning('请输入分类名称')
    return
  }
  categorySubmitting.value = true
  try {
    if (editingCategoryId.value) {
      const existing = flatCategories.value.find(c => c.id === editingCategoryId.value)
      const res = await updateCategory(editingCategoryId.value, { name: categoryForm.name, sort_order: existing?.sort_order ?? 0 })
      if (res.code === 0) {
        ElMessage.success('修改成功')
        showCategoryDialog.value = false
        fetchCategories()
      } else {
        ElMessage.error(res.message || '修改失败')
      }
    } else {
      const data: any = { name: categoryForm.name, sort_order: 9999 }
      if (categoryForm.parent_id) data.parent_id = categoryForm.parent_id
      const res = await createCategory(data)
      if (res.code === 0) {
        ElMessage.success('创建成功')
        showCategoryDialog.value = false
        fetchCategories()
      } else {
        ElMessage.error(res.message || '创建失败')
      }
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '操作失败')
  } finally {
    categorySubmitting.value = false
  }
}

async function handleDeleteCategory(id: number) {
  try {
    await ElMessageBox.confirm('删除分类将同时删除其下所有子分类，确定？', '确认删除', { type: 'warning' })
    const res = await deleteCategory(id)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      if (selectedCategoryId.value === id) {
        selectedCategoryId.value = null
      }
      fetchCategories()
      fetchDocuments()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch {
    // cancelled
  }
}

function toggleCategoryCheck(id: number, checked: boolean) {
  const newSet = new Set(checkedCategoryIds.value)
  if (checked) {
    newSet.add(id)
  } else {
    newSet.delete(id)
  }
  checkedCategoryIds.value = newSet
}

function collectAllChildIds(cat: Category): number[] {
  const ids: number[] = [cat.id]
  if (cat.children) {
    for (const child of cat.children) {
      ids.push(...collectAllChildIds(child))
    }
  }
  return ids
}

async function handleCascadeDelete() {
  const ids = Array.from(checkedCategoryIds.value)
  if (ids.length === 0) return

  // 统计涉及的文档数
  let totalDocs = 0
  const collectDocs = (cats: Category[]) => {
    for (const cat of cats) {
      if (checkedCategoryIds.value.has(cat.id)) {
        totalDocs += cat.doc_count ?? 0
      }
      if (cat.children) collectDocs(cat.children)
    }
  }
  collectDocs(categories.value)

  try {
    await ElMessageBox.confirm(
      `即将删除 ${ids.length} 个目录及其下所有子目录和 ${totalDocs} 篇文档，此操作不可恢复！确定？`,
      '⚠️ 级联删除确认',
      { type: 'error', confirmButtonText: '确认删除', cancelButtonText: '取消' }
    )
  } catch {
    return
  }

  try {
    const res = await cascadeDeleteCategories(ids)
    if (res.code === 0) {
      ElMessage.success(`已删除 ${res.data.deleted_categories} 个目录和 ${res.data.deleted_documents} 篇文档`)
      checkedCategoryIds.value = new Set()
      if (selectedCategoryId.value && ids.includes(selectedCategoryId.value)) {
        selectedCategoryId.value = null
      }
      fetchCategories()
      fetchDocuments()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch {
    ElMessage.error('删除失败')
  }
}

// ==================== 文档操作 ====================

function initTopSortable() {
  if (!treeListRef.value) return
  if (topSortable) {
    topSortable.destroy()
  }
  topSortable = Sortable.create(treeListRef.value, {
    handle: '.drag-handle',
    animation: 200,
    ghostClass: 'sortable-ghost',
    chosenClass: 'sortable-chosen',
    filter: '.root-node',
    onEnd(evt) {
      const oldIdx = evt.oldIndex
      const newIdx = evt.newIndex
      if (oldIdx == null || newIdx == null || oldIdx === newIdx) return
      // root-node is first child, so subtract 1
      const from: number = oldIdx - 1
      const to: number = newIdx - 1
      if (from < 0 || to < 0 || from >= categories.value.length) return
      const moved = categories.value.splice(from, 1)[0]
      categories.value.splice(to, 0, moved)
      saveCategoryOrder()
    },
  })
}

function handleReorderChildren(parentId: number, oldIndex: number, newIndex: number) {
  const parent = findInTree(categories.value, parentId)
  if (!parent?.children || oldIndex === newIndex) return
  const moved = parent.children.splice(oldIndex, 1)[0]
  parent.children.splice(newIndex, 0, moved)
  saveCategoryOrder()
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

async function saveCategoryOrder() {
  const ids = collectOrderedIds(categories.value)
  const items = ids.map((id, index) => ({ id, sort_order: index }))
  sortSaving.value = true
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
    sortSaving.value = false
  }
}

async function fetchDocuments() {
  docsLoading.value = true
  try {
    const params: any = { page: page.value, size: pageSize }
    if (selectedCategoryId.value !== null) {
      params.category_id = selectedCategoryId.value
    }
    if (statusFilter.value === 'published') {
      params.status = 'published'
    } else if (statusFilter.value === 'draft') {
      params.status = 'draft'
    }
    const res = await listDocuments(params)
    if (res.code === 0) {
      documents.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } catch {
    ElMessage.error('获取文档列表失败')
  } finally {
    docsLoading.value = false
  }
}

function handleSelectionChange(rows: Document[]) {
  selectedDocIds.value = rows.map(r => r.id)
}

function handleRowClick(row: Document) {
  openEditor(row)
}

async function handleCreateDoc() {
  try {
    const data: any = { title: '无标题文档', content: '' }
    if (selectedCategoryId.value !== null) {
      data.category_id = selectedCategoryId.value
    }
    const res = await createDocument(data)
    if (res.code === 0 && res.data) {
      ElMessage.success('创建成功')
      openEditor(res.data)
    } else {
      ElMessage.error(res.message || '创建失败')
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '创建失败')
  }
}

async function openEditor(doc: Document) {
  editingDoc.value = true
  editingDocId.value = doc.id
  saveStatus.value = ''
  try {
    const res = await getAdminDocument(doc.id)
    if (res.code === 0 && res.data) {
      editForm.title = res.data.title
      editForm.content = res.data.content
      editForm.category_id = res.data.category_id
      editForm.is_published = res.data.is_published
    }
  } catch {
    ElMessage.error('加载文档失败')
  }
}

function backToList() {
  editingDoc.value = false
  editingDocId.value = null
  fetchDocuments()
  fetchCategories() // 刷新分类的 doc_count
}

async function handleSave(isPublished: boolean) {
  if (!editForm.title.trim()) {
    ElMessage.warning('请输入标题')
    return
  }
  saving.value = true
  saveStatus.value = '保存中...'
  try {
    if (editingDocId.value) {
      const res = await updateDocument(editingDocId.value, {
        title: editForm.title,
        content: editForm.content,
        category_id: editForm.category_id || undefined,
      })
      if (res.code === 0) {
        if (isPublished) {
          await publishDocument(editingDocId.value, true)
          editForm.is_published = true
        }
        saveStatus.value = '已保存'
        ElMessage.success('保存成功')
      } else {
        saveStatus.value = '保存失败'
        ElMessage.error(res.message || '保存失败')
      }
    }
  } catch (err: any) {
    saveStatus.value = '保存失败'
    ElMessage.error(err.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handlePublish(row: Document) {
  const newState = !row.is_published
  const label = newState ? '发布' : '取消发布'
  try {
    await ElMessageBox.confirm(`确定${label}该文档？`, '确认操作', { type: 'warning' })
    const res = await publishDocument(row.id, newState)
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

async function handleDeleteDoc(id: number) {
  try {
    await ElMessageBox.confirm('确定删除该文档？此操作不可恢复。', '确认删除', { type: 'warning' })
    const res = await deleteDocument(id)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      fetchDocuments()
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
      `确定删除选中的 ${selectedDocIds.value.length} 篇文档？此操作不可恢复。`,
      '批量删除',
      { type: 'warning' }
    )
    const res = await batchDeleteDocuments(selectedDocIds.value)
    if (res.code === 0) {
      ElMessage.success(`成功删除 ${res.data?.deleted || 0} 篇文档`)
      selectedDocIds.value = []
      fetchDocuments()
      fetchCategories()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch {
    // cancelled
  }
}

async function handleUpload(options: any) {
  try {
    const res = await uploadFile(options.file)
    if (res.code === 0 && res.data) {
      const textarea = document.querySelector('.editor-textarea') as HTMLTextAreaElement
      if (textarea) {
        const start = textarea.selectionStart
        const imgMarkdown = `![${options.file.name}](${res.data.url})`
        editForm.content = editForm.content.slice(0, start) + imgMarkdown + editForm.content.slice(start)
      }
      ElMessage.success('上传成功')
    } else {
      ElMessage.error(res.message || '上传失败')
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '上传失败')
  }
}

function formatTime(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

// ==================== Lifecycle ====================

onMounted(() => {
  fetchCategories()
  fetchDocuments()
})

onBeforeUnmount(() => {
  if (topSortable) {
    topSortable.destroy()
    topSortable = null
  }
})

// 监听筛选变化
watch([statusFilter, docSearch], () => {
  page.value = 1
  fetchDocuments()
})

// 监听编辑表单变化
let autoSaveTimer: number | null = null
watch(
  () => editForm.content,
  () => {
    if (editingDoc.value) {
      saveStatus.value = '未保存'
      if (autoSaveTimer) clearTimeout(autoSaveTimer)
      autoSaveTimer = window.setTimeout(() => {
        saveStatus.value = '有未保存的更改'
      }, 2000)
    }
  }
)
</script>

<style lang="scss" scoped>
$color-bg: #F5F5F7;
$color-border: #E8E8ED;
$color-text-primary: #1D1D1F;
$color-text-secondary: #86868B;
$color-accent: #0071E3;
$color-hover: #F0F0F2;

.documents-manager {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 56px - 64px);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-shrink: 0;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.back-btn {
  font-size: 18px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.manager-body {
  display: flex;
  flex: 1;
  gap: 0;
  border: 1px solid $color-border;
  border-radius: 12px;
  overflow: hidden;
  background: white;
  min-height: 0;
  position: relative;
}

/* ====== 左侧分类面板 ====== */

.category-panel {
  width: 280px;
  min-width: 280px;
  border-right: 1px solid $color-border;
  display: flex;
  flex-direction: column;
  background: $color-bg;
}

.panel-header {
  padding: 16px;
  border-bottom: 1px solid $color-border;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.panel-title {
  font-size: 13px;
  font-weight: 600;
  color: $color-text-secondary;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.search-input {
  :deep(.el-input__wrapper) {
    border-radius: 8px;
    box-shadow: none;
    border: 1px solid $color-border;
  }
}

.tree-container {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.root-node {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  color: $color-text-primary;
  transition: background 0.15s;

  &:hover {
    background: $color-hover;
  }

  &.active {
    background: white;
    color: $color-accent;
    border-right: 3px solid $color-accent;
  }

  .node-name {
    flex: 1;
  }

  .node-count {
    font-size: 12px;
    color: $color-text-secondary;
    background: white;
    padding: 1px 8px;
    border-radius: 10px;
    font-weight: 500;
  }
}

.empty-tree {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 16px;
  gap: 12px;

  p {
    margin: 0;
    color: $color-text-secondary;
    font-size: 14px;
  }
}

/* ====== 右侧内容面板 ====== */

.content-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
  position: relative;
  z-index: 1;
}

/* -- 列表视图 -- */

.list-view {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

.list-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid $color-border;
  flex-shrink: 0;
}

.toolbar-left {
  .doc-count-label {
    font-size: 14px;
    color: $color-text-secondary;
  }
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.doc-search {
  width: 200px;
  :deep(.el-input__wrapper) {
    border-radius: 8px;
  }
}

.status-filter {
  width: 120px;
}

.doc-table {
  flex: 1;
  min-height: 0;
  overflow: hidden;

  :deep(.el-table) {
    height: 100%;
  }

  :deep(.el-table__body-wrapper) {
    overflow-y: auto;
  }

  :deep(.el-table__header) {
    th {
      background: $color-bg;
    }
  }
}

.doc-row {
  cursor: pointer;
}

.doc-title {
  font-weight: 500;
  color: $color-text-primary;
}

.doc-category {
  font-size: 13px;
  color: $color-text-secondary;
}

.doc-time {
  font-size: 13px;
  color: $color-text-secondary;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 16px 20px;
  border-top: 1px solid $color-border;
  flex-shrink: 0;
}

/* -- 编辑视图 -- */

.editor-view {
  display: flex;
  flex-direction: column;
  flex: 1;
  padding: 20px;
  min-height: 0;
}

.edit-meta {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  flex-shrink: 0;
}

.title-input {
  flex: 1;
  :deep(.el-input__wrapper) {
    border-radius: 8px;
  }
}

.category-select {
  width: 200px;
}

.editor-body {
  display: flex;
  flex: 1;
  gap: 1px;
  background: $color-border;
  border-radius: 8px;
  overflow: hidden;
  min-height: 0;
}

.editor-pane,
.preview-pane {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: white;
}

.pane-header {
  padding: 8px 16px;
  font-size: 12px;
  color: $color-text-secondary;
  background: $color-bg;
  border-bottom: 1px solid $color-border;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  flex-shrink: 0;
}

.editor-textarea {
  flex: 1;
  padding: 16px;
  border: none;
  resize: none;
  font-family: 'SF Mono', 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 14px;
  line-height: 1.6;
  color: $color-text-primary;
  outline: none;
  min-height: 0;
}

.upload-area {
  padding: 8px 16px;
  border-top: 1px solid #F2F2F7;
  flex-shrink: 0;
}

.preview-content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  min-height: 0;
}

.edit-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  flex-shrink: 0;
}

.footer-left {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: $color-text-secondary;
}

.footer-right {
  display: flex;
  gap: 8px;
}

.unsaved {
  color: #E6A23C;
}

/* ====== Responsive ====== */

@media (max-width: 900px) {
  .category-panel {
    width: 200px;
    min-width: 200px;
  }
}

@media (max-width: 700px) {
  .manager-body {
    flex-direction: column;
  }

  .category-panel {
    width: 100%;
    min-width: 100%;
    max-height: 200px;
    border-right: none;
    border-bottom: 1px solid $color-border;
  }

  .editor-body {
    flex-direction: column;
  }
}

/* 批量删除操作栏 */
.batch-delete-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  border-top: 1px solid $color-border;
  background: #FEF0F0;
  flex-shrink: 0;
}

.checked-count {
  font-size: 13px;
  color: #F56C6C;
  font-weight: 500;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.form-hint {
  font-size: 12px;
  color: #86868B;
  margin-top: 4px;
}
</style>
