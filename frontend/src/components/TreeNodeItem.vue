<template>
  <div class="tree-node-item" :class="[`depth-${level}`, { 'is-last-child': isLastChild }]">
    <!-- 连接线层 -->
    <div class="tree-lines" aria-hidden="true">
      <span v-for="i in level" :key="i" class="indent-line" :class="{ 'line-active': i < level }" />
    </div>

    <!-- 节点主体 -->
    <div
      class="tree-node"
      :class="{ expanded: treeStore.isExpanded(node.id) }"
      @click="handleClick"
    >
      <!-- 展开箭头 -->
      <span
        v-if="hasChildren || hasDocs"
        class="expand-icon"
        :class="{ expanded: treeStore.isExpanded(node.id) }"
        @click.stop="treeStore.toggleExpand(node.id)"
      >
        <el-icon :size="10"><ArrowRight /></el-icon>
      </span>
      <span v-else class="expand-placeholder" />

      <!-- 文件夹图标 -->
      <el-icon :size="15" class="node-icon folder-icon">
        <FolderOpened v-if="treeStore.isExpanded(node.id) && (hasChildren || hasDocs)" />
        <Folder v-else />
      </el-icon>

      <!-- 名称 -->
      <span class="node-name">{{ node.name }}</span>

      <!-- 文档数量徽章 -->
      <span v-if="totalDocCount > 0" class="doc-badge">{{ totalDocCount }}</span>
    </div>

    <!-- 子节点（grid 动画） -->
    <div
      v-if="hasChildren || hasDocs"
      class="tree-children-wrapper"
      :class="{ open: treeStore.isExpanded(node.id) }"
    >
      <div class="tree-children-inner">
        <!-- 子分类 -->
        <TreeNodeItem
          v-for="(child, idx) in node.children"
          :key="child.id"
          :node="child"
          :level="level + 1"
          :is-last-child="idx === (node.children?.length ?? 1) - 1 && (!node.docs || node.docs.length === 0)"
        />

        <!-- 分类下的文档 -->
        <div
          v-for="(doc, idx) in node.docs"
          :key="doc.id"
          class="tree-doc"
          :class="{
            active: currentDocId === doc.id,
            'is-last-child': idx === (node.docs?.length ?? 1) - 1,
            [`depth-${level + 1}`]: true,
          }"
          :style="{ '--indent': level + 1 }"
          @click="goToDoc(doc.id)"
        >
          <!-- 文档连接线 -->
          <span v-for="i in level + 1" :key="i" class="indent-line" :class="{ 'line-active': i < level + 1 }" />
          <el-icon :size="13" class="node-icon doc-icon"><Document /></el-icon>
          <span class="doc-name">{{ doc.title }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTreeStore } from '@/stores/tree'
import type { TreeNode } from '@/types'

const props = defineProps<{
  node: TreeNode
  level: number
  isLastChild?: boolean
}>()

const route = useRoute()
const router = useRouter()
const treeStore = useTreeStore()

const hasChildren = computed(() => (props.node.children?.length ?? 0) > 0)
const hasDocs = computed(() => (props.node.docs?.length ?? 0) > 0)
const currentDocId = computed(() => treeStore.activeDocId || Number(route.params.id) || 0)

// 当 activeDocId 变化时，滚动到当前文档位置
watch(
  () => treeStore.activeDocId,
  async (docId) => {
    if (!docId) return
    // 检查当前节点下是否包含该文档
    const hasDoc = props.node.docs?.some(d => d.id === docId)
    if (hasDoc) {
      await nextTick()
      // 延迟一下等展开动画完成
      setTimeout(() => {
        const el = document.querySelector(`.tree-doc.active`)
        if (el) {
          el.scrollIntoView({ behavior: 'smooth', block: 'center' })
        }
      }, 300)
    }
  }
)

const totalDocCount = computed(() => {
  let count = props.node.docs?.length ?? 0
  const countChildren = (nodes?: TreeNode[]) => {
    if (!nodes) return
    for (const n of nodes) {
      count += n.docs?.length ?? 0
      countChildren(n.children)
    }
  }
  countChildren(props.node.children)
  return count
})

function handleClick() {
  if (hasChildren.value || hasDocs.value) {
    treeStore.toggleExpand(props.node.id)
  }
  if (props.node.docs && props.node.docs.length > 0) {
    goToDoc(props.node.docs[0].id)
  }
}

function goToDoc(id: number) {
  router.push(`/admin/doc/${id}`)
}
</script>

<style lang="scss" scoped>
/* ─────── 颜色体系 ─────── */
$accent: #0071E3;
$text-0: #1D1D1F;
$text-1: #48484A;
$text-2: #636366;
$text-3: #86868B;
$text-4: #AEAEB2;
$bg-hover: rgba(0, 0, 0, 0.03);
$line-color: #D1D1D6;
$badge-bg: #F2F2F7;

/* ─────── 层级颜色递减 ─────── */
$depth-colors: (
  0: $text-0,
  1: $text-1,
  2: $text-2,
  3: $text-3,
);

/* ─────── 根容器 ─────── */
.tree-node-item {
  position: relative;
}

/* ─────── 连接线层 ─────── */
.tree-lines {
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  pointer-events: none;
  display: flex;
}

.indent-line {
  width: 20px;
  position: relative;
  flex-shrink: 0;

  &::before {
    content: '';
    position: absolute;
    left: 9px;
    top: 0;
    bottom: 0;
    width: 1px;
    background: $line-color;
    opacity: 0;
    transition: opacity 0.15s;
  }

  &.line-active::before {
    opacity: 1;
  }
}

/* 最后一个子节点：竖线只到中点 */
.tree-node-item.is-last-child > .tree-lines > .indent-line:last-child {
  &::before {
    bottom: 50%;
  }
}

/* ─────── 节点主体 ─────── */
.tree-node {
  position: relative;
  display: flex;
  align-items: center;
  height: 32px;
  padding: 0 12px 0 0;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  gap: 4px;
  user-select: none;
  transition: background 0.18s ease;

  /* 缩进 */
  padding-left: calc(12px + var(--indent, 0) * 0px);
  margin-left: calc(var(--indent, 0) * 0px);

  &:hover {
    background: $bg-hover;
  }

  &.expanded > .node-icon.folder-icon {
    color: $accent;
  }
}

/* 每个节点的缩进通过 depth-N 类控制 */
@for $i from 0 through 6 {
  .depth-#{$i} > .tree-node,
  .depth-#{$i} > .tree-doc {
    padding-left: #{$i * 20 + 8}px;
  }
}

/* ─────── 展开箭头 ─────── */
.expand-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 4px;
  flex-shrink: 0;
  color: $text-3;
  transition: transform 0.2s ease, color 0.15s;

  &.expanded {
    transform: rotate(90deg);
    color: $accent;
  }

  &:hover {
    background: rgba(0, 0, 0, 0.06);
  }
}

.expand-placeholder {
  width: 16px;
  flex-shrink: 0;
}

/* ─────── 图标 ─────── */
.node-icon {
  flex-shrink: 0;
  transition: color 0.15s;
}

.folder-icon {
  color: #C0C4CC;
}

/* 层级颜色递减 */
@each $level, $color in $depth-colors {
  .depth-#{$level} > .tree-node .node-name,
  .depth-#{$level} > .tree-doc .doc-name {
    color: $color;
  }
}

/* ─────── 名称 ─────── */
.node-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 32px;
}

/* ─────── 文档数量徽章 ─────── */
.doc-badge {
  font-size: 11px;
  font-weight: 500;
  color: $text-3;
  background: $badge-bg;
  padding: 0 6px;
  border-radius: 8px;
  height: 18px;
  line-height: 18px;
  flex-shrink: 0;
  margin-left: 4px;
}

/* ─────── 文档行 ─────── */
.tree-doc {
  position: relative;
  display: flex;
  align-items: center;
  height: 30px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  gap: 4px;
  transition: background 0.18s ease;
  padding-right: 12px;

  &:hover {
    background: $bg-hover;
  }

  &.active {
    .doc-name {
      color: $accent;
      font-weight: 600;
    }
    .doc-icon {
      color: $accent;
    }
  }

  .doc-icon {
    color: $text-4;
    flex-shrink: 0;
  }

  .doc-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    line-height: 30px;
    color: $text-2;
    transition: color 0.15s;
  }

  &.is-last-child > .indent-line:last-child::before {
    bottom: 50%;
  }
}

/* ─────── 子节点容器（grid 动画） ─────── */
.tree-children-wrapper {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows 0.22s ease;

  &.open {
    grid-template-rows: 1fr;
  }
}

.tree-children-inner {
  overflow: hidden;
  min-height: 0;
}
</style>
