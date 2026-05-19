<template>
  <div class="category-tree-node">
    <div
      class="tree-node"
      :class="{ active: selectedId === category.id, [`level-${level}`]: true }"
      @click="$emit('select', category.id)"
      @contextmenu.prevent="showContextMenu"
    >
      <el-checkbox
        :model-value="checkedIds.has(category.id)"
        class="node-checkbox"
        @change="(val: boolean) => $emit('toggle-check', category.id, val)"
        @click.stop
      />

      <span class="drag-handle" title="拖拽排序">
        <svg width="12" height="12" viewBox="0 0 14 14" fill="currentColor">
          <circle cx="4.5" cy="2.5" r="1.3" />
          <circle cx="9.5" cy="2.5" r="1.3" />
          <circle cx="4.5" cy="7" r="1.3" />
          <circle cx="9.5" cy="7" r="1.3" />
          <circle cx="4.5" cy="11.5" r="1.3" />
          <circle cx="9.5" cy="11.5" r="1.3" />
        </svg>
      </span>

      <span
        v-if="hasChildren"
        class="expand-icon"
        :class="{ expanded: expanded }"
        @click.stop="expanded = !expanded"
      >
        <el-icon :size="12"><ArrowRight /></el-icon>
      </span>
      <span v-else class="expand-placeholder" />

      <el-icon :size="14" class="folder-icon"><Folder /></el-icon>
      <span class="node-name">{{ category.name }}</span>
      <span class="node-count">{{ category.doc_count ?? 0 }}</span>
    </div>

    <!-- 子分类 -->
    <transition name="expand">
      <div v-if="hasChildren && expanded" ref="childrenRef" class="tree-children">
        <CategoryTreeNode
          v-for="child in category.children"
          :key="child.id"
          :category="child"
          :level="level + 1"
          :selected-id="selectedId"
          :search-text="searchText"
          :checked-ids="checkedIds"
          @select="(id: number) => $emit('select', id)"
          @edit="(cat: Category) => $emit('edit', cat)"
          @delete="(id: number) => $emit('delete', id)"
          @add-child="(id: number) => $emit('add-child', id)"
          @toggle-check="(id: number, val: boolean) => $emit('toggle-check', id, val)"
          @reorder-children="(parentId: number, oldIdx: number, newIdx: number) => $emit('reorderChildren', parentId, oldIdx, newIdx)"
        />
      </div>
    </transition>

    <!-- 右键菜单 -->
    <Teleport to="body">
      <div
        v-if="contextMenu.visible"
        class="context-menu"
        :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
        @click="contextMenu.visible = false"
      >
        <div class="menu-item" @click="$emit('select', category.id)">
          <el-icon><View /></el-icon>
          <span>查看文档</span>
        </div>
        <div class="menu-item" @click="$emit('add-child', category.id)">
          <el-icon><FolderAdd /></el-icon>
          <span>添加子分类</span>
        </div>
        <div class="menu-item" @click="$emit('edit', category)">
          <el-icon><Edit /></el-icon>
          <span>重命名</span>
        </div>
        <div class="menu-divider" />
        <div class="menu-item danger" @click="$emit('delete', category.id)">
          <el-icon><Delete /></el-icon>
          <span>删除</span>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import Sortable from 'sortablejs'
import type { Category } from '@/types'

const props = defineProps<{
  category: Category
  level: number
  selectedId: number | null
  searchText: string
  checkedIds: Set<number>
}>()

const emit = defineEmits<{
  select: [id: number]
  edit: [cat: Category]
  delete: [id: number]
  'add-child': [id: number]
  'toggle-check': [id: number, val: boolean]
  reorderChildren: [parentId: number, oldIndex: number, newIndex: number]
}>()

const expanded = ref(true)
const childrenRef = ref<HTMLElement | null>(null)
let childrenSortable: Sortable | null = null

const hasChildren = computed(() => props.category.children && props.category.children.length > 0)

// 搜索时自动展开
if (props.searchText) {
  expanded.value = true
}

function initChildrenSortable() {
  if (!childrenRef.value) return
  if (childrenSortable) {
    childrenSortable.destroy()
  }
  childrenSortable = Sortable.create(childrenRef.value, {
    handle: '.drag-handle',
    animation: 200,
    ghostClass: 'sortable-ghost',
    chosenClass: 'sortable-chosen',
    onEnd(evt) {
      const oldIdx = evt.oldIndex
      const newIdx = evt.newIndex
      if (oldIdx == null || newIdx == null || oldIdx === newIdx) return
      emit('reorderChildren', props.category.id, oldIdx as number, newIdx as number)
    },
  })
}

function destroyChildrenSortable() {
  if (childrenSortable) {
    childrenSortable.destroy()
    childrenSortable = null
  }
}

watch(expanded, async (val) => {
  if (val && props.category.children?.length) {
    await nextTick()
    initChildrenSortable()
  } else if (!val) {
    destroyChildrenSortable()
  }
})

onMounted(async () => {
  if (expanded.value && props.category.children?.length) {
    await nextTick()
    initChildrenSortable()
  }
})

onBeforeUnmount(() => {
  destroyChildrenSortable()
  if (typeof window !== 'undefined') {
    window.removeEventListener('click', hideContextMenu)
  }
})

// 右键菜单
const contextMenu = ref({ visible: false, x: 0, y: 0 })

function showContextMenu(e: MouseEvent) {
  contextMenu.value = {
    visible: true,
    x: e.clientX,
    y: e.clientY,
  }
}

function hideContextMenu() {
  contextMenu.value.visible = false
}

// 点击任意位置关闭菜单
if (typeof window !== 'undefined') {
  window.addEventListener('click', hideContextMenu)
}
</script>

<style lang="scss" scoped>
$color-text-primary: #1D1D1F;
$color-text-secondary: #86868B;
$color-bg-hover: #F0F0F2;
$color-accent: #0071E3;

.tree-node {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  cursor: pointer;
  color: $color-text-primary;
  font-size: 14px;
  transition: background 0.15s ease;
  gap: 6px;
  user-select: none;

  &:hover {
    background: $color-bg-hover;
  }

  &.active {
    background: white;
    color: $color-accent;
    font-weight: 500;
    border-right: 3px solid $color-accent;
  }

  &.level-0 { padding-left: 16px; }
  &.level-1 { padding-left: 36px; }
  &.level-2 { padding-left: 56px; }
  &.level-3 { padding-left: 76px; }
}

.node-checkbox {
  flex-shrink: 0;
  margin-right: 2px;
}

.drag-handle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  cursor: grab;
  color: #D1D1D6;
  flex-shrink: 0;
  transition: color 0.15s ease;
  user-select: none;

  &:hover {
    color: #8E8E93;
  }

  &:active {
    cursor: grabbing;
    color: #636366;
  }
}

.expand-icon {
  display: inline-flex;
  transition: transform 0.2s ease;
  color: $color-text-secondary;
  flex-shrink: 0;
  &.expanded {
    transform: rotate(90deg);
  }
}

.expand-placeholder {
  width: 12px;
  display: inline-block;
  flex-shrink: 0;
}

.folder-icon {
  color: #C0C4CC;
  flex-shrink: 0;
}

.node-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-count {
  font-size: 12px;
  color: $color-text-secondary;
  background: #F2F2F7;
  padding: 1px 8px;
  border-radius: 10px;
  flex-shrink: 0;
}

.expand-enter-active,
.expand-leave-active {
  transition: all 0.2s ease;
  overflow: hidden;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
}

.expand-enter-to,
.expand-leave-from {
  opacity: 1;
  max-height: 500px;
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

<style lang="scss">
/* 右键菜单 - 全局样式 */
.context-menu {
  position: fixed;
  z-index: 9999;
  background: white;
  border: 1px solid #E8E8ED;
  border-radius: 8px;
  padding: 4px 0;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  min-width: 160px;

  .menu-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    font-size: 14px;
    color: #1D1D1F;
    cursor: pointer;
    transition: background 0.15s;

    &:hover {
      background: #F0F0F2;
    }

    &.danger {
      color: #F56C6C;
      &:hover {
        background: #FEF0F0;
      }
    }
  }

  .menu-divider {
    height: 1px;
    background: #E8E8ED;
    margin: 4px 0;
  }
}
</style>
