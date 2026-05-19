<template>
  <div class="document-page">
    <div v-if="loading" class="loading-state">
      <el-icon class="is-loading" :size="24"><Loading /></el-icon>
    </div>

    <template v-else-if="doc">
      <h1 class="doc-title">{{ doc.title }}</h1>
      <div class="doc-meta">
        <span v-if="doc.author_name">{{ doc.author_name }}</span>
        <span v-if="doc.updated_at">· {{ formatDate(doc.updated_at) }}</span>
      </div>
      <div class="doc-content">
        <MarkdownViewer :content="doc.content" :title="doc.title" />
      </div>
    </template>

    <div v-else class="empty-state">
      文档不存在
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getDocument } from '@/api/bff'
import { useTreeStore } from '@/stores/tree'
import MarkdownViewer from '@/components/MarkdownViewer.vue'
import type { Document } from '@/types'

const route = useRoute()
const treeStore = useTreeStore()
const doc = ref<Document | null>(null)
const loading = ref(false)

async function loadDoc(id: number) {
  // 如果已有文档内容，先不显示 loading，保持当前内容
  if (!doc.value) {
    loading.value = true
  }
  try {
    const res = await getDocument(id)
    if (res.code === 0) {
      doc.value = res.data
      // 加载文档后，自动展开左侧目录树到对应分类
      syncTreeToDoc(doc.value, id)
    } else {
      doc.value = null
    }
  } catch {
    doc.value = null
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })
}

/**
 * 同步左侧目录树：展开到文档所在分类并高亮
 * 优先使用 URL 中的 cat 参数，其次使用文档自身的 category_id
 */
function syncTreeToDoc(document: Document, docId: number) {
  const catFromQuery = Number(route.query.cat)
  const categoryId = catFromQuery || document.category_id
  if (categoryId) {
    treeStore.expandToCategory(categoryId, docId)
  } else {
    treeStore.activeDocId = docId
  }
}

onMounted(() => {
  loadDoc(Number(route.params.id))
})

watch(
  () => route.params.id,
  (id) => {
    if (id) loadDoc(Number(id))
  }
)

// 监听 URL cat 参数变化（从搜索结果跳转时）
watch(
  () => route.query.cat,
  () => {
    if (doc.value) {
      syncTreeToDoc(doc.value, Number(route.params.id))
    }
  }
)

// 监听树数据加载完成（树可能在文档之后才加载完成）
watch(
  () => treeStore.tree,
  (newTree) => {
    if (newTree.length > 0 && doc.value) {
      syncTreeToDoc(doc.value, Number(route.params.id))
    }
  }
)
</script>

<style lang="scss" scoped>
.document-page {
  max-width: 800px;
  margin: 0 auto;
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 64px;
  color: #AEAEB2;
}

.empty-state {
  text-align: center;
  padding: 64px;
  color: #AEAEB2;
  font-size: 16px;
}

.doc-title {
  font-size: 36px;
  font-weight: 600;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', sans-serif;
  color: #1D1D1F;
  margin-bottom: 12px;
  line-height: 1.2;
}

.doc-meta {
  font-size: 14px;
  color: #86868B;
  margin-bottom: 32px;
  padding-bottom: 16px;
  border-bottom: 1px solid #F2F2F7;
}

.doc-content {
  padding-bottom: 64px;
}
</style>
