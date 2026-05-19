<template>
  <div
    class="category-row"
    :style="{ paddingLeft: level * 24 + 16 + 'px' }"
  >
    <div class="row-content">
      <el-checkbox
        :model-value="selectedIds.includes(category.id)"
        @change="$emit('toggleSelect', category.id)"
        @click.stop
      />

      <span class="drag-handle" title="拖拽排序">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="currentColor">
          <circle cx="4.5" cy="2.5" r="1.3" />
          <circle cx="9.5" cy="2.5" r="1.3" />
          <circle cx="4.5" cy="7" r="1.3" />
          <circle cx="9.5" cy="7" r="1.3" />
          <circle cx="4.5" cy="11.5" r="1.3" />
          <circle cx="9.5" cy="11.5" r="1.3" />
        </svg>
      </span>

      <span
        v-if="category.children && category.children.length > 0"
        class="expand-icon"
        :class="{ expanded }"
        @click="expanded = !expanded"
      >
        <el-icon :size="12"><ArrowRight /></el-icon>
      </span>
      <span v-else class="expand-placeholder" />

      <span class="cat-name">{{ category.name }}</span>
      <span class="doc-count">{{ category.doc_count ?? 0 }}</span>

      <div class="row-actions">
        <el-button type="primary" link size="small" @click="$emit('addChild', category.id)">
          添加子分类
        </el-button>
        <el-button type="primary" link size="small" @click="$emit('edit', category)">
          编辑
        </el-button>
        <el-button type="danger" link size="small" @click="$emit('delete', category.id)">
          删除
        </el-button>
      </div>
    </div>

    <transition name="expand">
      <div
        v-if="expanded && category.children && category.children.length > 0"
        ref="childrenRef"
        class="sortable-children"
      >
        <CategoryRow
          v-for="child in category.children"
          :key="child.id"
          :category="child"
          :level="level + 1"
          :selected-ids="selectedIds"
          @edit="(c: any) => $emit('edit', c)"
          @delete="(id: number) => $emit('delete', id)"
          @add-child="(id: number) => $emit('addChild', id)"
          @toggle-select="(id: number) => $emit('toggleSelect', id)"
          @reorder-children="(parentId: number, oldIdx: number, newIdx: number) => $emit('reorderChildren', parentId, oldIdx, newIdx)"
        />
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import Sortable from 'sortablejs'
import type { Category } from '@/types'

const props = defineProps<{
  category: Category
  level: number
  selectedIds: number[]
}>()

const emit = defineEmits<{
  edit: [cat: Category]
  delete: [id: number]
  addChild: [id: number]
  toggleSelect: [id: number]
  reorderChildren: [parentId: number, oldIndex: number, newIndex: number]
}>()

const expanded = ref(true)
const childrenRef = ref<HTMLElement | null>(null)
let childrenSortable: Sortable | null = null

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
      // 通过 emit 冒泡到 Categories.vue 处理
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

// 监听展开状态变化
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
})
</script>

<style lang="scss" scoped>
.category-row {
  border-bottom: 1px solid #F2F2F7;
  &:last-child {
    border-bottom: none;
  }
}

.row-content {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  gap: 8px;
}

.drag-handle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  cursor: grab;
  color: #C7C7CC;
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
  cursor: pointer;
  transition: transform 0.2s ease;
  color: #86868B;
  &.expanded {
    transform: rotate(90deg);
  }
}

.expand-placeholder {
  width: 12px;
  display: inline-block;
}

.cat-name {
  flex: 1;
  font-size: 14px;
  color: #1D1D1F;
}

.doc-count {
  width: 60px;
  text-align: center;
  font-size: 13px;
  color: #86868B;
}

.row-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.15s ease;
}

.category-row:hover .row-actions {
  opacity: 1;
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
