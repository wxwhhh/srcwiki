<template>
  <div class="doc-edit-page">
    <div class="edit-header">
      <div class="edit-header-left">
        <el-button text @click="router.push('/admin/dashboard/documents')">
          <el-icon><Back /></el-icon>
          返回列表
        </el-button>
      </div>
      <div class="edit-header-right">
        <el-button :loading="saving" @click="handleSave(false)">
          保存草稿
        </el-button>
        <el-button type="primary" :loading="saving" @click="handleSave(true)">
          发布
        </el-button>
      </div>
    </div>

    <div class="edit-meta">
      <el-input
        v-model="form.title"
        placeholder="文档标题"
        class="title-input"
        size="large"
      />
      <el-select
        v-model="form.category_id"
        placeholder="选择分类"
        clearable
        class="category-select"
      >
        <el-option
          v-for="cat in flatCategories"
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
          v-model="form.content"
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
          <MarkdownViewer :content="form.content" />
        </div>
      </div>
    </div>

    <div class="edit-footer">
      <span class="word-count">字数：{{ wordCount }}</span>
      <span class="save-status">{{ saveStatus }}</span>
    </div>

    <!-- 版本历史 -->
    <div class="versions-section" v-if="docId !== 'new'">
      <h3 class="section-title">版本历史</h3>
      <el-table :data="versions" style="width: 100%" v-loading="versionsLoading">
        <el-table-column prop="id" label="版本ID" width="80" />
        <el-table-column prop="title" label="标题" />
        <el-table-column label="编辑者" width="120">
          <template #default="{ row }">
            {{ row.editor_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleRollback(row.id)">
              回滚
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getAdminDocument,
  updateDocument,
  publishDocument,
  listCategories,
  getDocumentVersions,
  rollbackDocument,
  uploadFile,
} from '@/api/admin'
import MarkdownViewer from '@/components/MarkdownViewer.vue'
import type { Category, DocumentVersion } from '@/types'

const route = useRoute()
const router = useRouter()
const docId = computed(() => route.params.id as string)

const saving = ref(false)
const saveStatus = ref('')
const versions = ref<DocumentVersion[]>([])
const versionsLoading = ref(false)

const form = reactive({
  title: '',
  content: '',
  category_id: null as number | null,
})

const flatCategories = ref<{ id: number; label: string }[]>([])

const wordCount = computed(() => {
  return form.content.replace(/\s/g, '').length
})

async function loadDocument() {
  if (docId.value === 'new') return
  try {
    const res = await getAdminDocument(Number(docId.value))
    if (res.code === 0 && res.data) {
      form.title = res.data.title
      form.content = res.data.content
      form.category_id = res.data.category_id
    }
  } catch {
    ElMessage.error('加载文档失败')
  }
}

async function loadCategories() {
  try {
    const res = await listCategories()
    if (res.code === 0) {
      flatCategories.value = flattenCategories(res.data || [])
    }
  } catch {
    // ignore
  }
}

function flattenCategories(cats: Category[], prefix = ''): { id: number; label: string }[] {
  const result: { id: number; label: string }[] = []
  for (const cat of cats) {
    const label = prefix ? `${prefix} / ${cat.name}` : cat.name
    result.push({ id: cat.id, label })
    if (cat.children && cat.children.length > 0) {
      result.push(...flattenCategories(cat.children, label))
    }
  }
  return result
}

async function loadVersions() {
  if (docId.value === 'new') return
  versionsLoading.value = true
  try {
    const res = await getDocumentVersions(Number(docId.value))
    if (res.code === 0) {
      versions.value = res.data || []
    }
  } catch {
    // ignore
  } finally {
    versionsLoading.value = false
  }
}

async function handleSave(isPublished: boolean) {
  if (!form.title.trim()) {
    ElMessage.warning('请输入标题')
    return
  }

  saving.value = true
  saveStatus.value = '保存中...'
  try {
    if (docId.value === 'new') {
      // 创建新文档
      const { createDocument } = await import('@/api/admin')
      const res = await createDocument({
        title: form.title,
        content: form.content,
        category_id: form.category_id || undefined,
      })
      if (res.code === 0 && res.data) {
        if (isPublished) {
          await publishDocument(res.data.id, true)
        }
        ElMessage.success('保存成功')
        router.replace(`/admin/dashboard/documents/${res.data.id}`)
      }
    } else {
      const res = await updateDocument(Number(docId.value), {
        title: form.title,
        content: form.content,
        category_id: form.category_id || undefined,
      })
      if (res.code === 0) {
        if (isPublished) {
          await publishDocument(Number(docId.value), true)
        }
        saveStatus.value = '已保存'
        ElMessage.success('保存成功')
        loadVersions()
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

async function handleRollback(vid: number) {
  try {
    await ElMessageBox.confirm('确定回滚到该版本？当前内容将被覆盖。', '确认回滚', { type: 'warning' })
    const res = await rollbackDocument(Number(docId.value), vid)
    if (res.code === 0) {
      ElMessage.success('回滚成功')
      loadDocument()
      loadVersions()
    } else {
      ElMessage.error(res.message || '回滚失败')
    }
  } catch {
    // cancelled
  }
}

async function handleUpload(options: any) {
  try {
    const res = await uploadFile(options.file)
    if (res.code === 0 && res.data) {
      // 在光标位置插入图片
      const textarea = document.querySelector('.editor-textarea') as HTMLTextAreaElement
      if (textarea) {
        const start = textarea.selectionStart
        const imgMarkdown = `![${options.file.name}](${res.data.url})`
        form.content = form.content.slice(0, start) + imgMarkdown + form.content.slice(start)
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

onMounted(() => {
  loadDocument()
  loadCategories()
  loadVersions()
})

// 自动保存提示
let autoSaveTimer: number | null = null
watch(
  () => form.content,
  () => {
    saveStatus.value = '未保存'
    if (autoSaveTimer) clearTimeout(autoSaveTimer)
    autoSaveTimer = window.setTimeout(() => {
      saveStatus.value = '有未保存的更改'
    }, 2000)
  }
)
</script>

<style lang="scss" scoped>
.doc-edit-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 56px - 64px);
}

.edit-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.edit-meta {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
}

.title-input {
  flex: 1;
}

.category-select {
  width: 200px;
}

.editor-body {
  display: flex;
  flex: 1;
  gap: 1px;
  background: #E8E8ED;
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
  color: #86868B;
  background: #F5F5F7;
  border-bottom: 1px solid #E8E8ED;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.editor-textarea {
  flex: 1;
  padding: 16px;
  border: none;
  resize: none;
  font-family: 'SF Mono', 'JetBrains Mono', monospace;
  font-size: 14px;
  line-height: 1.6;
  color: #1D1D1F;
  outline: none;
}

.upload-area {
  padding: 8px 16px;
  border-top: 1px solid #F2F2F7;
}

.preview-content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}

.edit-footer {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  font-size: 13px;
  color: #86868B;
}

.versions-section {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid #F2F2F7;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
}
</style>
