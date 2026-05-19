<template>
  <div class="search-page">
    <h1 class="page-title">搜索结果</h1>
    <p class="search-query" v-if="query">
      关键词：<strong>{{ query }}</strong>
    </p>

    <div v-if="loading" class="loading-state">
      <el-icon class="is-loading" :size="24"><Loading /></el-icon>
    </div>

    <div v-else-if="results.length === 0 && query" class="empty-state">
      未找到相关文档
    </div>

    <div v-else class="result-list">
      <div
        v-for="hit in results"
        :key="hit.id"
        class="result-card"
        @click="goToDoc(hit)"
      >
        <h3 class="result-title" v-html="highlightText(hit.title)" />
        <p class="result-snippet" v-html="highlightText(hit.content_snippet)" />
        <div class="result-meta">
          <span v-if="hit.category_name">{{ hit.category_name }}</span>
          <span v-if="hit.updated_at">· {{ formatDate(hit.updated_at) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { search } from '@/api/bff'
import type { SearchHit } from '@/types'

const route = useRoute()
const router = useRouter()

const query = ref('')
const results = ref<SearchHit[]>([])
const loading = ref(false)

function highlightText(text: string): string {
  if (!query.value || !text) return text
  // 先转义 HTML 实体，防止 XSS
  const escapedText = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
  // 再做关键词高亮
  const escapedQuery = query.value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return escapedText.replace(new RegExp(`(${escapedQuery})`, 'gi'), '<mark>$1</mark>')
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })
}

function goToDoc(hit: SearchHit) {
  const catParam = hit.category_id ? `?cat=${hit.category_id}` : ''
  router.push(`/admin/doc/${hit.id}${catParam}`)
}

async function doSearch(q: string) {
  if (!q) {
    results.value = []
    return
  }
  loading.value = true
  try {
    const res = await search(q)
    if (res.code === 0) {
      results.value = res.data.list || []
    }
  } catch {
    results.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  query.value = (route.query.q as string) || ''
  if (query.value) doSearch(query.value)
})

watch(
  () => route.query.q,
  (q) => {
    query.value = (q as string) || ''
    if (query.value) doSearch(query.value)
  }
)
</script>

<style lang="scss" scoped>
.search-page {
  max-width: 800px;
  margin: 0 auto;
}

.page-title {
  font-size: 36px;
  font-weight: 600;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', sans-serif;
  margin-bottom: 8px;
}

.search-query {
  font-size: 16px;
  color: #86868B;
  margin-bottom: 32px;
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

.result-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.result-card {
  padding: 24px;
  border-radius: 12px;
  cursor: pointer;
  transition: background 0.15s ease;
  border: 1px solid #F2F2F7;

  &:hover {
    background: #F5F5F7;
  }
}

.result-title {
  font-size: 18px;
  font-weight: 600;
  color: #1D1D1F;
  margin-bottom: 8px;

  :deep(mark) {
    background: #FFE066;
    padding: 0 2px;
    border-radius: 2px;
  }
}

.result-snippet {
  font-size: 14px;
  color: #86868B;
  line-height: 1.6;
  margin-bottom: 8px;

  :deep(mark) {
    background: #FFE066;
    padding: 0 2px;
    border-radius: 2px;
  }
}

.result-meta {
  font-size: 13px;
  color: #AEAEB2;
}
</style>
