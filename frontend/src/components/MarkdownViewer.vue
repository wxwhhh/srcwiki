<template>
  <div class="markdown-body" ref="rootEl">
    <!-- 渲染内容 -->
    <div v-html="sanitizedHtml" />

    <!-- Lightbox 遮罩 -->
    <Teleport to="body">
      <Transition name="lightbox">
        <div v-if="lightboxSrc" class="lightbox-overlay" @click="closeLightbox">
          <img :src="lightboxSrc" class="lightbox-img" alt="" />
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'

// ─── Props ───
const props = defineProps<{
  content: string
  title?: string
}>()

// ─── 状态 ───
const lightboxSrc = ref<string | null>(null)
const rootEl = ref<HTMLElement | null>(null)

// ─── 配置 marked ───
marked.setOptions({ breaks: true, gfm: true })

// slug 去重
const slugCounts = new Map<string, number>()

function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/<[^>]*>/g, '')
    .replace(/[^\w\u4e00-\u9fff]+/g, '-')
    .replace(/^-+|-+$/g, '')
    || 'heading'
}

function uniqueSlug(base: string): string {
  const count = slugCounts.get(base) ?? 0
  slugCounts.set(base, count + 1)
  return count === 0 ? base : `${base}-${count}`
}

// 自定义渲染器
const renderer = new marked.Renderer()

// 代码块 — macOS Terminal 风格 + 行号
renderer.code = function ({ text, lang }: { text: string; lang?: string }) {
  const language = lang && hljs.getLanguage(lang) ? lang : 'plaintext'
  const highlighted = hljs.highlight(text, { language }).value
  const lines = highlighted.split('\n')
  // 最后一行可能是空的（trailing newline），去掉
  if (lines.length > 1 && lines[lines.length - 1].trim() === '') lines.pop()
  const lineCount = lines.length
  const gutterWidth = String(lineCount).length
  const lineNums = lines.map((_, i) => {
    const num = String(i + 1).padStart(gutterWidth, ' ')
    return `<span class="line-num">${num}</span>`
  }).join('\n')
  const codeHtml = lines.join('\n')

  return `<div class="code-block" data-lang="${language}"><div class="code-titlebar"><div class="traffic-lights"><span class="tl-red"></span><span class="tl-yellow"></span><span class="tl-green"></span></div><span class="code-lang">${language}</span><button class="copy-btn">复制</button></div><div class="code-body"><div class="code-gutter">${lineNums}</div><pre class="code-pre"><code class="hljs language-${language}">${codeHtml}</code></pre></div></div>`
}

// 标题 + 锚点
renderer.heading = function ({ text, depth }: { text: string; depth: number }) {
  const plainText = text.replace(/<[^>]*>/g, '')
  const slug = uniqueSlug(slugify(plainText))
  const anchor = depth >= 2 ? `<a class="heading-anchor" href="#${slug}" aria-label="链接">#</a>` : ''
  return `<h${depth} id="${slug}">${anchor}${text}</h${depth}>`
}

// 图片加 lightbox 属性
renderer.image = function ({ href, title, text }: { href: string; title: string | null; text: string }) {
  const titleAttr = title ? ` title="${title}"` : ''
  return `<img src="${href}" alt="${text}"${titleAttr} class="zoomable-img" loading="lazy" />`
}

marked.use({ renderer })

// ─── 标准化标题 ───
function normalizeTitle(s: string): string {
  return s.replace(/[\s\-_.,;:!?()\[\]{}"'`~@#$%^&*+=|\\/<>]/g, '').toLowerCase()
}

// ─── 渲染 HTML ───
const sanitizedHtml = computed(() => {
  if (!props.content) return ''
  let content = props.content.replace(/^\uFEFF/, '')

  if (props.title) {
    const titleNorm = normalizeTitle(props.title)
    let changed = true
    while (changed) {
      changed = false
      content = content.replace(/^[ \t]*\r?\n/, '')
      const m = content.match(/^#([^#\n][^\n]*)\r?\n?/)
      if (m) {
        const h1Text = m[1].trim()
        const h1Norm = normalizeTitle(h1Text)
        if (h1Norm === titleNorm || titleNorm.includes(h1Norm) || h1Norm.includes(titleNorm)) {
          content = content.slice(m[0].length)
          changed = true
        }
      }
    }
  } else {
    content = content.replace(/^[ \t]*#[^#].*\r?\n?/, '')
  }
  content = content.trimStart()

  slugCounts.clear()
  const rawHtml = marked.parse(content) as string

  return DOMPurify.sanitize(rawHtml, {
    ALLOWED_TAGS: [
      'p', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
      'ul', 'ol', 'li', 'blockquote', 'pre', 'code',
      'table', 'thead', 'tbody', 'tr', 'th', 'td',
      'a', 'img', 'strong', 'em', 'br', 'hr', 'div', 'span',
      'del', 'input', 'sup', 'sub', 'button',
    ],
    ALLOWED_ATTR: [
      'href', 'src', 'alt', 'class', 'target', 'rel',
      'type', 'checked', 'disabled', 'colspan', 'rowspan',
      'id', 'title', 'aria-label', 'data-lang',
    ],
  })
})

// ─── 事件委托（在 mounted 后绑定到 root 元素）───
function handleClick(e: Event) {
  const target = e.target as HTMLElement

  // 代码复制
  if (target.classList.contains('copy-btn')) {
    const codeEl = target.closest('.code-block')?.querySelector('code')
    if (codeEl) {
      const text = codeEl.textContent ?? ''
      // 优先用 clipboard API，失败则用 execCommand 兜底
      const doCopy = (text: string) => {
        if (navigator.clipboard && window.isSecureContext) {
          return navigator.clipboard.writeText(text)
        }
        // execCommand 兜底（HTTP 环境）
        const ta = document.createElement('textarea')
        ta.value = text
        ta.style.cssText = 'position:fixed;left:-9999px;top:-9999px'
        document.body.appendChild(ta)
        ta.select()
        document.execCommand('copy')
        document.body.removeChild(ta)
        return Promise.resolve()
      }
      doCopy(text).then(() => {
        target.textContent = '已复制 ✓'
        setTimeout(() => { target.textContent = '复制' }, 2000)
      }).catch(() => {
        target.textContent = '复制失败'
        setTimeout(() => { target.textContent = '复制' }, 2000)
      })
    }
    return
  }

  // 图片放大
  if (target.classList.contains('zoomable-img')) {
    lightboxSrc.value = (target as HTMLImageElement).src
  }
}

// ─── Lightbox ───
function closeLightbox() {
  lightboxSrc.value = null
}

function onEsc(e: KeyboardEvent) {
  if (e.key === 'Escape' && lightboxSrc.value) {
    closeLightbox()
  }
}

// ─── 生命周期 ───
onMounted(() => {
  rootEl.value?.addEventListener('click', handleClick)
  document.addEventListener('keydown', onEsc)
})

onUnmounted(() => {
  rootEl.value?.removeEventListener('click', handleClick)
  document.removeEventListener('keydown', onEsc)
})
</script>

<style lang="scss" scoped>
$accent: #0071E3;
$text-2: #86868B;
$text-3: #AEAEB2;

/* ─────── Code Block — macOS Terminal ─────── */
:deep(.code-block) {
  margin: 20px 0;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #E5E5EA;
  background: #FAFAFA;
}

:deep(.code-titlebar) {
  display: flex;
  align-items: center;
  padding: 8px 14px;
  background: #F2F2F7;
  border-bottom: 1px solid #E5E5EA;
  gap: 8px;
}

:deep(.traffic-lights) {
  display: flex;
  gap: 6px;
  flex-shrink: 0;

  span {
    width: 12px;
    height: 12px;
    border-radius: 50%;
  }
}

:deep(.tl-red)    { background: #FF5F56; }
:deep(.tl-yellow) { background: #FFBD2E; }
:deep(.tl-green)  { background: #27C93F; }

:deep(.code-lang) {
  flex: 1;
  text-align: center;
  font-size: 12px;
  color: #86868B;
  letter-spacing: 0.5px;
  user-select: none;
}

:deep(.copy-btn) {
  font-size: 12px;
  color: #86868B;
  background: #FFFFFF;
  border: 1px solid #D1D1D6;
  border-radius: 4px;
  padding: 2px 10px;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
  opacity: 0;

  &:hover {
    background: #F2F2F7;
    color: #1D1D1F;
  }
}

:deep(.code-block:hover .copy-btn) {
  opacity: 1;
}

:deep(.code-body) {
  display: flex;
  background: #FFFFFF;
}

:deep(.code-gutter) {
  flex-shrink: 0;
  padding: 16px 0;
  background: #FAFAFA;
  border-right: 1px solid #E5E5EA;
  user-select: none;
  text-align: right;
}

:deep(.line-num) {
  display: block;
  padding: 0 12px;
  font-family: 'SF Mono', 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #AEAEB2;
}

:deep(.code-pre) {
  flex: 1;
  margin: 0;
  padding: 16px;
  overflow-x: auto;
  background: transparent;

  code {
    font-family: 'SF Mono', 'JetBrains Mono', 'Fira Code', monospace;
    font-size: 13px;
    line-height: 1.6;
    background: none;
    padding: 0;
  }
}

/* ─────── Heading Anchor ─────── */
:deep(.heading-anchor) {
  display: none;
  float: left;
  margin-left: -1.2em;
  padding-right: 0.4em;
  color: $text-3;
  font-weight: 400;
  text-decoration: none;
  transition: color 0.15s;

  &:hover {
    color: $accent;
  }
}

:deep(h2:hover .heading-anchor),
:deep(h3:hover .heading-anchor),
:deep(h4:hover .heading-anchor),
:deep(h5:hover .heading-anchor),
:deep(h6:hover .heading-anchor) {
  display: inline;
}

/* ─────── 图片可点击 ─────── */
:deep(.zoomable-img) {
  cursor: zoom-in;
  transition: opacity 0.15s;

  &:hover {
    opacity: 0.85;
  }
}

/* ─────── Lightbox ─────── */
.lightbox-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: zoom-out;
}

.lightbox-img {
  max-width: 90vw;
  max-height: 90vh;
  object-fit: contain;
  border-radius: 8px;
}

.lightbox-enter-active,
.lightbox-leave-active {
  transition: opacity 0.2s ease;
}

.lightbox-enter-from,
.lightbox-leave-to {
  opacity: 0;
}
</style>
