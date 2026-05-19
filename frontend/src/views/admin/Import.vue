<template>
  <div class="import-page">
    <div class="page-header">
      <h1 class="page-title">批量导入</h1>
    </div>

    <!-- Tab 切换 -->
    <div class="tab-bar">
      <button class="tab-btn" :class="{ active: activeTab === 'zip' }" @click="activeTab = 'zip'">
        ZIP 文件导入
      </button>
      <button class="tab-btn" :class="{ active: activeTab === 'github' }" @click="activeTab = 'github'">
        GitHub 仓库导入
      </button>
    </div>

    <!-- ZIP 导入 Tab -->
    <div v-if="activeTab === 'zip'" class="tab-content">
      <div class="import-card">
        <div class="upload-area" :class="{ dragover }" @dragover.prevent="dragover = true" @dragleave="dragover = false" @drop.prevent="handleDrop">
          <el-icon :size="48" color="#C7C7CC"><UploadFilled /></el-icon>
          <p class="upload-title">拖拽 ZIP 文件到此处</p>
          <p class="upload-hint">或点击下方按钮选择文件（支持多选）</p>
          <input ref="fileInput" type="file" accept=".zip" multiple style="display:none" @change="handleFileSelect" />
          <el-button type="primary" @click="fileInput?.click()" :loading="uploadProgress.uploading" :disabled="uploadProgress.uploading">
            <el-icon><Upload /></el-icon>
            选择 ZIP 文件
          </el-button>
        </div>

        <div v-if="uploadProgress.uploading" class="upload-progress">
          <el-progress :percentage="100" status="success" :format="() => `正在上传 ${uploadProgress.total} 个文件...`" />
          <p class="upload-status">正在上传并创建导入任务，请稍候...</p>
        </div>

        <div class="import-rules">
          <h4>导入规则</h4>
          <ul>
            <li>📁 文件夹 → 分类节点（层级自动保留）</li>
            <li>📄 <code>.md</code> 文件 → 文档（文件名 = 标题）</li>
            <li>🖼️ 图片文件 → 自动提取并替换文档中的引用路径</li>
            <li>🔄 <strong>自动去重</strong>：同名分类跳过，同标题文档跳过</li>
            <li>🔤 <strong>编码自动检测</strong>：GBK/GB2312/Big5 自动转 UTF-8</li>
            <li>📦 ZIP 最大 500MB，单个 .md 最大 1MB</li>
          </ul>
          <p class="async-hint">💡 导入将在后台异步执行，上传后自动跳转到任务列表页面查看进度。</p>
        </div>
      </div>
    </div>

    <!-- GitHub 导入 Tab -->
    <div v-if="activeTab === 'github'" class="tab-content">
      <div class="import-card">
        <div class="github-form">
          <div class="form-row">
            <div class="form-field">
              <label class="field-label">仓库 URL</label>
              <el-input
                v-model="githubUrl"
                placeholder="https://github.com/owner/repo"
                size="large"
                :disabled="githubImporting"
              />
            </div>
          </div>
          <div class="form-row">
            <div class="form-field half">
              <label class="field-label">分支名</label>
              <el-input
                v-model="githubBranch"
                placeholder="main"
                size="large"
                :disabled="githubImporting"
              />
            </div>
            <div class="form-field half">
              <label class="field-label">&nbsp;</label>
              <el-checkbox v-model="githubSkipRoot" :disabled="githubImporting">
                跳过仓库根目录
              </el-checkbox>
            </div>
          </div>
          <div class="form-row">
            <div class="form-field">
              <el-button
                type="primary"
                size="large"
                :loading="githubImporting"
                :disabled="!githubUrl.trim()"
                class="import-btn"
                @click="startGithubImport"
              >
                <el-icon><Download /></el-icon>
                开始导入
              </el-button>
            </div>
          </div>
        </div>

        <div class="import-rules">
          <h4>导入规则</h4>
          <ul>
            <li>📁 目录结构 → 分类层级（自动创建嵌套分类）</li>
            <li>📄 <code>.md</code> 文件 → 文档（文件名 = 标题）</li>
            <li>🔄 <strong>自动去重</strong>：同标题文档跳过</li>
            <li>🚫 跳过隐藏文件/目录（<code>.git</code>、<code>.github</code> 等）</li>
            <li>📦 单个 .md 最大 1MB</li>
            <li>⏱️ 仓库克隆超时 5 分钟</li>
          </ul>
          <p class="async-hint">💡 导入将在后台异步执行，提交后自动跳转到任务列表页面查看进度。</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createBatchZipImportTask, createGithubImportTask } from '@/api/admin'

const router = useRouter()

// ─── Tab ───
const activeTab = ref<'zip' | 'github'>('zip')

// ─── ZIP 导入 ───
const fileInput = ref<HTMLInputElement>()
const dragover = ref(false)
const uploadProgress = ref({ current: 0, total: 0, uploading: false })

async function uploadMultiple(files: File[]) {
  const zipFiles = files.filter(f => f.name.endsWith('.zip'))
  if (zipFiles.length === 0) {
    ElMessage.warning('请上传 .zip 文件')
    return
  }

  uploadProgress.value = { current: zipFiles.length, total: zipFiles.length, uploading: true }

  try {
    const res = await createBatchZipImportTask(zipFiles)
    if (res.code === 0) {
      ElMessage.success(`已创建 ${res.data.count} 个导入任务`)
      router.push('/admin/dashboard/tasks')
    } else {
      ElMessage.error(res.message || '批量上传失败')
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '批量上传失败')
  } finally {
    uploadProgress.value.uploading = false
  }
}

function handleFileSelect(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files?.length) {
    uploadMultiple(Array.from(input.files))
  }
  input.value = ''
}

function handleDrop(e: DragEvent) {
  dragover.value = false
  const files = e.dataTransfer?.files
  if (files?.length) {
    uploadMultiple(Array.from(files))
  }
}

// ─── GitHub 导入 ───
const githubUrl = ref('')
const githubBranch = ref('main')
const githubSkipRoot = ref(false)
const githubImporting = ref(false)

async function startGithubImport() {
  if (!githubUrl.value.trim()) {
    ElMessage.warning('请输入仓库 URL')
    return
  }

  githubImporting.value = true

  try {
    const res = await createGithubImportTask({
      url: githubUrl.value.trim(),
      branch: githubBranch.value.trim(),
      skip_root: githubSkipRoot.value,
    })
    if (res.code === 0) {
      ElMessage.success('GitHub 导入任务已创建，正在跳转...')
      setTimeout(() => router.push('/admin/dashboard/tasks'), 1000)
    } else {
      ElMessage.error(res.message || '创建任务失败')
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '创建任务失败')
  } finally {
    githubImporting.value = false
  }
}
</script>

<style lang="scss" scoped>
.page-header {
  margin-bottom: 16px;
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', sans-serif;
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

  &:hover { color: #1D1D1F; }
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

/* ─── 共用样式 ─── */
.import-card {
  background: white;
  border: 1px solid #E8E8ED;
  border-radius: 12px;
  padding: 32px;
}

.upload-area {
  border: 2px dashed #D1D1D6;
  border-radius: 12px;
  padding: 48px 32px;
  text-align: center;
  transition: all 0.2s ease;
  cursor: pointer;

  &.dragover { border-color: #0071E3; background: #F0F7FF; }
  &:hover { border-color: #AEAEB2; }
}

.upload-title { font-size: 16px; font-weight: 500; color: #1D1D1F; margin: 12px 0 4px; }
.upload-hint { font-size: 13px; color: #86868B; margin-bottom: 16px; }

/* ─── GitHub 表单 ─── */
.github-form {
  margin-bottom: 8px;
}

.form-row {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  align-items: flex-end;
}

.form-field {
  flex: 1;

  &.half { flex: 0 0 calc(50% - 8px); }
}

.field-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: #1D1D1F;
  margin-bottom: 6px;
}

.import-btn {
  height: 40px;
  min-width: 120px;
}

/* ─── 导入规则 ─── */
.import-rules {
  margin-top: 24px;
  padding: 20px 24px;
  background: #F5F5F7;
  border-radius: 8px;

  h4 { font-size: 14px; font-weight: 600; color: #1D1D1F; margin-bottom: 10px; }
  ul { list-style: none; padding: 0; margin: 0; }
  li { font-size: 13px; color: #424245; padding: 4px 0; }
  code { background: #E8E8ED; padding: 1px 5px; border-radius: 4px; font-size: 12px; }
}

.async-hint {
  margin-top: 12px;
  padding: 10px 14px;
  background: #E8F4FD;
  border-radius: 6px;
  font-size: 13px;
  color: #1565C0;
}

/* ─── 上传进度 ─── */
.upload-progress {
  margin-bottom: 24px;
  padding: 20px 24px;
  background: #F5F5F7;
  border-radius: 8px;
}

.upload-status {
  margin-top: 8px;
  font-size: 13px;
  color: #86868B;
  text-align: center;
}
</style>
