<template>
  <div class="home-page">
    <!-- ① 标题区 -->
    <section class="hero">
      <h1 class="hero-title">LiteWiki 知识库</h1>
      <p class="hero-subtitle">轻量级漏洞知识管理平台</p>
    </section>

    <!-- ② 免责声明 -->
    <section class="disclaimer">
      <p class="disclaimer-text">本系统仅供学习研究使用，请遵守相关法律法规。</p>
    </section>

    <!-- ③ 统计卡片 -->
    <section class="stats">
      <div class="stat-card">
        <div class="stat-item">
          <span class="stat-icon">📄</span>
          <span class="stat-label">文档</span>
          <span class="stat-number">{{ stats.total_docs }}</span>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-item">
          <span class="stat-icon">📁</span>
          <span class="stat-label">分类</span>
          <span class="stat-number">{{ stats.total_cats }}</span>
        </div>
      </div>
    </section>

    <!-- ④ 最近更新 -->
    <section v-if="stats.recent_docs.length > 0" class="recent">
      <div class="section-header">
        <h2 class="section-title">最近更新</h2>
      </div>
      <div class="recent-list">
        <div
          v-for="doc in stats.recent_docs"
          :key="doc.id"
          class="recent-row"
          @click="router.push(`/admin/doc/${doc.id}`)"
        >
          <span class="recent-title">{{ doc.title }}</span>
          <span class="recent-cat">{{ doc.category_name }}</span>
          <span class="recent-time">{{ relativeTime(doc.updated_at) }}</span>
        </div>
      </div>
    </section>

    <!-- ⑤ 数据来源 -->
    <section v-if="credits.length > 0" class="sources">
      <div class="section-header">
        <h2 class="section-title">数据来源</h2>
      </div>
      <div class="sources-grid">
        <a
          v-for="credit in credits"
          :key="credit.id"
          :href="credit.url"
          target="_blank"
          rel="noopener"
          class="source-item"
        >
          <img v-if="credit.icon_url" :src="credit.icon_url" class="source-icon" alt="" />
          <span class="source-name">{{ credit.name }}</span>
        </a>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTreeStore } from '@/stores/tree'
import { listDocsByCategory, getCredits, getHomeStats } from '@/api/bff'
import type { Document, Credit, HomeStats } from '@/types'

const route = useRoute()
const router = useRouter()
const treeStore = useTreeStore()

const credits = ref<Credit[]>([])

const stats = reactive<HomeStats>({
  total_docs: 0,
  total_cats: 0,
  recent_docs: [],
})

onMounted(() => {
  treeStore.fetchTree()
  fetchCredits()
  fetchHomeStats()
})

async function fetchHomeStats() {
  try {
    const res = await getHomeStats()
    if (res.code === 0 && res.data) {
      stats.total_docs = res.data.total_docs
      stats.total_cats = res.data.total_cats
      stats.recent_docs = (res.data.recent_docs || []).slice(0, 10)
    }
  } catch {
    // silent
  }
}

async function fetchCredits() {
  try {
    const res = await getCredits()
    if (res.code === 0) {
      credits.value = res.data || []
    }
  } catch {
    // silent
  }
}

function relativeTime(dateStr: string): string {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diff = now - then
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)
  const months = Math.floor(days / 30)

  if (seconds < 60) return '刚刚'
  if (minutes < 60) return `${minutes} 分钟前`
  if (hours < 24) return `${hours} 小时前`
  if (days < 30) return `${days} 天前`
  if (months < 12) return `${months} 个月前`
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

// Keep category query compat (unused on new home but harmless)
watch(
  () => route.query.category,
  async (catId) => {
    if (catId) {
      await listDocsByCategory(Number(catId))
    }
  },
  { immediate: true }
)
</script>

<style lang="scss" scoped>
// ── Variables ──
$font-sans: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'PingFang SC', 'Helvetica Neue', sans-serif;
$font-text: -apple-system, BlinkMacSystemFont, 'SF Pro Text', 'PingFang SC', 'Helvetica Neue', sans-serif;

$bg-page: #FFFFFF;
$bg-glass: rgba(255, 255, 255, 0.72);
$bg-glass-hover: rgba(255, 255, 255, 0.88);
$bg-disclaimer: rgba(255, 248, 230, 0.85);
$border-disclaimer: rgba(255, 214, 10, 0.5);
$text-primary: #1D1D1F;
$text-secondary: #6E6E73;
$text-tertiary: #AEAEB2;
$border-glass: rgba(255, 255, 255, 0.18);
$border-subtle: rgba(0, 0, 0, 0.06);
$accent: #0071E3;
$shadow-card: 0 2px 20px rgba(0, 0, 0, 0.06), 0 0 1px rgba(0, 0, 0, 0.08);
$shadow-card-hover: 0 8px 40px rgba(0, 0, 0, 0.1), 0 0 1px rgba(0, 0, 0, 0.1);
$shadow-glass: 0 4px 30px rgba(0, 0, 0, 0.04);
$radius-lg: 16px;
$radius-md: 14px;
$radius-sm: 10px;

// ── Page ──
.home-page {
  max-width: 940px;
  margin: 0 auto;
  padding: 48px 0 80px;
  font-family: $font-text;
  color: $text-primary;
}

// ── ① Hero ──
.hero {
  text-align: center;
  margin-bottom: 40px;
  padding: 32px 0 0;
}

.hero-title {
  font-family: $font-sans;
  font-size: 36px;
  font-weight: 700;
  letter-spacing: -0.5px;
  color: $text-primary;
  margin: 0 0 10px;
  line-height: 1.2;
}

.hero-subtitle {
  font-size: 17px;
  font-weight: 400;
  color: $text-secondary;
  margin: 0;
  letter-spacing: 0.2px;
}

// ── ② Disclaimer ──
.disclaimer {
  margin: 0 auto 44px;
  max-width: 100%;
  padding: 16px 28px;
  background: $bg-disclaimer;
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid $border-disclaimer;
  border-radius: $radius-lg;
  box-shadow: $shadow-glass;
  text-align: center;
}

.disclaimer-text {
  font-size: 14px;
  font-weight: 400;
  color: #9A7B2E;
  line-height: 1.7;
  margin: 0;
  letter-spacing: 0.1px;
}

// ── ③ Stats ──
.stats {
  display: flex;
  justify-content: center;
  margin-bottom: 44px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 0;
  background: $bg-glass;
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border: 1px solid $border-glass;
  border-radius: $radius-lg;
  padding: 16px 32px;
  transition: all 0.35s cubic-bezier(0.25, 0.46, 0.45, 0.94);
  cursor: default;
  box-shadow: $shadow-card;

  &:hover {
    background: $bg-glass-hover;
    transform: translateY(-2px);
    box-shadow: $shadow-card-hover;
  }
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat-icon {
  font-size: 18px;
  opacity: 0.85;
}

.stat-label {
  font-size: 13px;
  font-weight: 500;
  color: $text-tertiary;
  letter-spacing: 0.5px;
}

.stat-number {
  font-family: $font-sans;
  font-size: 22px;
  font-weight: 700;
  color: $text-primary;
  line-height: 1;
  letter-spacing: -0.5px;
}

.stat-divider {
  width: 1px;
  height: 28px;
  background: $border-subtle;
  margin: 0 24px;
}

// ── Section Headers ──
.section-header {
  padding-bottom: 14px;
  border-bottom: 1px solid $border-subtle;
  margin-bottom: 4px;
}

.section-title {
  font-family: $font-sans;
  font-size: 14px;
  font-weight: 600;
  color: $text-secondary;
  text-transform: uppercase;
  letter-spacing: 1.5px;
  margin: 0;
}

// ── ④ 最近更新 ──
.recent {
  margin-bottom: 44px;
}

.recent-list {
  background: $bg-glass;
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border: 1px solid $border-glass;
  border-radius: $radius-lg;
  overflow: hidden;
  box-shadow: $shadow-card;
}

.recent-row {
  display: flex;
  align-items: center;
  padding: 16px 24px;
  cursor: pointer;
  transition: background 0.2s ease;
  border-bottom: 1px solid rgba(0, 0, 0, 0.04);

  &:last-child {
    border-bottom: none;
  }

  &:hover {
    background: rgba(0, 113, 227, 0.04);
  }
}

.recent-title {
  flex: 1;
  font-size: 15px;
  font-weight: 500;
  color: $text-primary;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.recent-cat {
  font-size: 12px;
  font-weight: 500;
  color: $accent;
  background: rgba(0, 113, 227, 0.08);
  padding: 3px 10px;
  border-radius: 6px;
  margin: 0 16px;
  white-space: nowrap;
  flex-shrink: 0;
}

.recent-time {
  font-size: 13px;
  color: $text-tertiary;
  white-space: nowrap;
  flex-shrink: 0;
}

// ── ⑤ 数据来源 ──
.sources {
  margin-bottom: 44px;
}

.sources-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 4px;
}

.source-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  background: $bg-glass;
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid $border-glass;
  border-radius: $radius-sm;
  text-decoration: none;
  transition: all 0.25s ease;
  font-size: 13px;
  font-weight: 500;
  color: $text-primary;
  box-shadow: 0 1px 8px rgba(0, 0, 0, 0.04);

  &:hover {
    background: $bg-glass-hover;
    transform: translateY(-1px);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
    text-decoration: none;
  }
}

.source-icon {
  width: 18px;
  height: 18px;
  border-radius: 4px;
  object-fit: contain;
  flex-shrink: 0;
}

.source-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

// ── Responsive ──
@media (max-width: 768px) {
  .home-page {
    padding: 32px 0 64px;
  }

  .hero-title {
    font-size: 28px;
  }

  .hero-subtitle {
    font-size: 15px;
  }

  .stat-card {
    padding: 14px 20px;
  }

  .stat-number {
    font-size: 18px;
  }

  .stat-divider {
    margin: 0 16px;
  }

  .recent-row {
    padding: 14px 16px;
    flex-wrap: wrap;
    gap: 6px;
  }

  .recent-cat {
    margin: 0 8px;
  }

  .recent-time {
    flex-basis: 100%;
    order: 3;
    margin-top: 2px;
    margin-left: 0;
  }
}
</style>
