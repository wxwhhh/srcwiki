import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { TreeNode } from '@/types'
import { getTree } from '@/api/bff'

export const useTreeStore = defineStore('tree', () => {
  const tree = ref<TreeNode[]>([])
  const loading = ref(false)
  const expandedIds = ref<Set<number>>(new Set())
  const activeDocId = ref<number>(0)

  async function fetchTree() {
    loading.value = true
    try {
      const res = await getTree()
      if (res.code === 0) {
        tree.value = res.data || []
      }
    } catch {
      tree.value = []
    } finally {
      loading.value = false
    }
  }

  function toggleExpand(id: number) {
    if (expandedIds.value.has(id)) {
      expandedIds.value.delete(id)
    } else {
      expandedIds.value.add(id)
    }
  }

  function isExpanded(id: number): boolean {
    return expandedIds.value.has(id)
  }

  /**
   * 展开到指定分类，并标记当前文档为活跃状态
   * @param categoryId 分类 ID
   * @param docId 文档 ID（可选，用于高亮）
   */
  function expandToCategory(categoryId: number, docId?: number) {
    // 找到从根到目标分类的路径
    const path = findPathToCategory(tree.value, categoryId)
    if (path) {
      for (const id of path) {
        expandedIds.value.add(id)
      }
    }
    if (docId) {
      activeDocId.value = docId
    }
  }

  /**
   * 在树中查找从根到指定分类的路径（返回所有需要展开的节点 ID）
   */
  function findPathToCategory(nodes: TreeNode[], targetId: number): number[] | null {
    for (const node of nodes) {
      if (node.id === targetId) {
        return [node.id]
      }
      if (node.children) {
        const childPath = findPathToCategory(node.children, targetId)
        if (childPath) {
          return [node.id, ...childPath]
        }
      }
    }
    return null
  }

  return { tree, loading, expandedIds, activeDocId, fetchTree, toggleExpand, isExpanded, expandToCategory }
})
