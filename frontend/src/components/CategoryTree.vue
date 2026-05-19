<template>
  <div class="category-tree">
    <div v-if="treeStore.loading" class="tree-loading">
      <el-icon class="is-loading"><Loading /></el-icon>
    </div>
    <div v-else-if="treeStore.tree.length === 0" class="tree-empty">
      <el-icon :size="28" color="#AEAEB2"><FolderOpened /></el-icon>
      <span>暂无分类</span>
    </div>
    <template v-else>
      <TreeNodeItem
        v-for="(node, idx) in treeStore.tree"
        :key="node.id"
        :node="node"
        :level="0"
        :is-last-child="idx === treeStore.tree.length - 1"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useTreeStore } from '@/stores/tree'
import TreeNodeItem from './TreeNodeItem.vue'

const treeStore = useTreeStore()

onMounted(() => {
  treeStore.fetchTree()
})
</script>

<style lang="scss" scoped>
.category-tree {
  padding: 6px 0;
}

.tree-loading {
  display: flex;
  justify-content: center;
  padding: 24px;
  color: #AEAEB2;
}

.tree-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 32px 16px;
  color: #AEAEB2;
  font-size: 13px;
}
</style>
