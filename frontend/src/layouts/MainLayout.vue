<template>
  <div class="main-layout">
    <!-- 顶部栏 -->
    <header class="topbar">
      <div class="topbar-left">
        <button class="sidebar-toggle" @click="sidebarOpen = !sidebarOpen">
          <el-icon :size="20"><Fold v-if="sidebarOpen" /><Expand v-else /></el-icon>
        </button>
      </div>
      <div class="topbar-center">
        <SearchBox />
      </div>
      <div class="topbar-right">
        <el-dropdown trigger="click" @command="handleCommand">
          <span class="user-trigger">
            {{ auth.user?.username }}
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人信息</el-dropdown-item>
              <el-dropdown-item v-if="auth.isAdmin" command="admin">管理后台</el-dropdown-item>
              <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <div class="main-body">
      <!-- 侧边栏 -->
      <aside class="sidebar" :class="{ open: sidebarOpen }">
        <div class="sidebar-logo" @click="router.push('/admin/')">
          <span class="logo-text">LiteWiki</span>
        </div>
        <div class="sidebar-content">
          <CategoryTree />
        </div>
      </aside>

      <!-- 移动端遮罩 -->
      <div
        v-if="sidebarOpen"
        class="sidebar-overlay"
        @click="sidebarOpen = false"
      />

      <!-- 内容区 -->
      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import CategoryTree from '@/components/CategoryTree.vue'
import SearchBox from '@/components/SearchBox.vue'

const router = useRouter()
const auth = useAuthStore()
const sidebarOpen = ref(true)

function handleCommand(cmd: string) {
  if (cmd === 'profile') {
    router.push('/admin/profile')
  } else if (cmd === 'admin') {
    router.push('/admin/dashboard')
  } else if (cmd === 'logout') {
    auth.logout()
  }
}
</script>

<style lang="scss" scoped>
$color-bg-secondary: #F5F5F7;
$color-border: #E8E8ED;
$color-text-primary: #1D1D1F;

.main-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.topbar {
  height: 56px;
  border-bottom: 1px solid $color-border;
  display: flex;
  align-items: center;
  padding: 0 24px;
  background: white;
  position: sticky;
  top: 0;
  z-index: 100;

  &-left {
    width: 200px;
    flex-shrink: 0;
  }

  &-center {
    flex: 1;
    display: flex;
    justify-content: center;
  }

  &-right {
    width: 200px;
    flex-shrink: 0;
    display: flex;
    justify-content: flex-end;
  }
}

.sidebar-toggle {
  display: none;
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  color: $color-text-primary;
}

.user-trigger {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  font-size: 14px;
  color: $color-text-primary;
}

.main-body {
  display: flex;
  flex: 1;
}

.sidebar {
  width: 260px;
  background: $color-bg-secondary;
  border-right: 1px solid $color-border;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  position: sticky;
  top: 56px;
  height: calc(100vh - 56px);
  overflow-y: auto;

  &-logo {
    padding: 20px 24px;
    cursor: pointer;
  }

  .logo-text {
    font-size: 20px;
    font-weight: 600;
    font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', sans-serif;
    color: $color-text-primary;
  }

  &-content {
    flex: 1;
    padding: 0 12px 24px;
    overflow-y: auto;
  }
}

.sidebar-overlay {
  display: none;
}

.content {
  margin: 0 auto;
  flex: 1;
  padding: 32px 48px;
  max-width: 1080px;
  min-width: 0;
}

@media (max-width: 768px) {
  .sidebar-toggle {
    display: block;
  }

  .topbar-left {
    width: auto;
  }

  .topbar-right {
    width: auto;
  }

  .sidebar {
    position: fixed;
    top: 56px;
    left: 0;
    bottom: 0;
    z-index: 200;
    transform: translateX(-100%);
    transition: transform 0.2s ease;

    &.open {
      transform: translateX(0);
    }
  }

  .sidebar-overlay {
    display: block;
    position: fixed;
    top: 56px;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.3);
    z-index: 199;
  }

  .content {
    padding: 24px 16px;
  }
}
</style>
