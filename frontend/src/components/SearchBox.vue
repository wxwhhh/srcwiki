<template>
  <div class="search-box-wrapper">
    <el-icon class="search-icon" :size="16"><Search /></el-icon>
    <input
      v-model="query"
      type="text"
      class="search-input"
      placeholder="搜索文档..."
      @keyup.enter="doSearch"
    />
    <button class="search-btn" @click="doSearch" :disabled="!query.trim()">
      <el-icon :size="16"><Search /></el-icon>
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const query = ref('')

function doSearch() {
  if (query.value.trim()) {
    router.push({ path: '/admin/search', query: { q: query.value.trim() } })
  }
}
</script>

<style lang="scss" scoped>
$search-bg: #F5F5F7;
$search-border: #E8E8ED;
$search-accent: #0071E3;

.search-box-wrapper {
  position: relative;
  width: 320px;
}

.search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: #AEAEB2;
  pointer-events: none;
}

.search-input {
  width: 100%;
  height: 40px;
  border: 1px solid $search-border;
  border-radius: 9999px;
  padding: 0 40px 0 40px;
  font-size: 14px;
  background: $search-bg;
  color: #1D1D1F;
  font-family: inherit;
  transition: all 0.2s ease;

  &::placeholder {
    color: #AEAEB2;
  }

  &:focus {
    outline: none;
    background: white;
    border-color: $search-accent;
    box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.1);
    width: 400px;
  }
}

.search-btn {
  position: absolute;
  right: 4px;
  top: 50%;
  transform: translateY(-50%);
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 50%;
  background: $search-accent;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
  padding: 0;

  &:hover:not(:disabled) {
    background: #005BBB;
    transform: translateY(-50%) scale(1.05);
  }

  &:active:not(:disabled) {
    transform: translateY(-50%) scale(0.95);
  }

  &:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
}

@media (max-width: 768px) {
  .search-box-wrapper {
    width: 200px;
  }

  .search-input:focus {
    width: 260px;
  }
}
</style>
