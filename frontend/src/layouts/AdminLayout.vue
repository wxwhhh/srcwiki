<template>
  <div class="admin-layout">
    <!-- 顶部栏 -->
    <header class="admin-topbar">
      <div class="topbar-left">
        <router-link to="/admin/" class="back-link">
          <el-icon><Back /></el-icon>
          <span>返回前台</span>
        </router-link>
      </div>
      <div class="topbar-center">
        <span class="admin-title">管理后台</span>
      </div>
      <div class="topbar-right">
        <span class="username">{{ auth.user?.username }}</span>
      </div>
    </header>

    <div class="admin-body">
      <!-- 侧边菜单 -->
      <aside class="admin-sidebar">
        <nav class="admin-nav">
          <router-link
            v-for="item in menuItems"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isActive(item.path) }"
          >
            <el-icon :size="18"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>
        </nav>
      </aside>

      <!-- 内容区 -->
      <main class="admin-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { onMounted } from 'vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

onMounted(async () => {
  if (!auth.user) {
    await auth.fetchUser()
  }
  if (!auth.isAdmin) {
    router.push('/admin/')
  }
})

const menuItems = [
  { path: '/admin/dashboard', label: '概览', icon: 'DataBoard' },
  { path: '/admin/dashboard/users', label: '用户管理', icon: 'User' },
  { path: '/admin/dashboard/invites', label: '邀请码', icon: 'Ticket' },
  { path: '/admin/dashboard/content', label: '内容管理', icon: 'Document' },
  { path: '/admin/dashboard/import', label: '批量导入', icon: 'Upload' },
  { path: '/admin/dashboard/tasks', label: '导入任务', icon: 'List' },
  { path: '/admin/dashboard/audit', label: '审计日志', icon: 'Document' },
  { path: '/admin/dashboard/credits', label: '致谢管理', icon: 'Star' },
  { path: '/admin/dashboard/settings', label: '系统设置', icon: 'Setting' },
]

function isActive(path: string): boolean {
  if (path === '/admin/dashboard') return route.path === '/admin/dashboard'
  return route.path.startsWith(path)
}
</script>

<style lang="scss" scoped>
$color-bg-secondary: #F5F5F7;
$color-border: #E8E8ED;
$color-text-primary: #1D1D1F;
$color-text-secondary: #86868B;
$color-accent: #0071E3;

.admin-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.admin-topbar {
  height: 56px;
  border-bottom: 1px solid $color-border;
  display: flex;
  align-items: center;
  padding: 0 24px;
  background: white;
  position: sticky;
  top: 0;
  z-index: 100;

  .topbar-left { width: 200px; }
  .topbar-center { flex: 1; text-align: center; }
  .topbar-right { width: 200px; text-align: right; }
}

.back-link {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: $color-accent;
  text-decoration: none;
  &:hover { text-decoration: underline; }
}

.admin-title {
  font-size: 16px;
  font-weight: 600;
  color: $color-text-primary;
}

.username {
  font-size: 14px;
  color: $color-text-secondary;
}

.admin-body {
  display: flex;
  flex: 1;
}

.admin-sidebar {
  width: 220px;
  background: $color-bg-secondary;
  border-right: 1px solid $color-border;
  padding: 16px 12px;
  position: sticky;
  top: 56px;
  height: calc(100vh - 56px);
}

.admin-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 14px;
  color: $color-text-primary;
  text-decoration: none;
  transition: background 0.15s ease;

  &:hover {
    background: #F0F0F2;
    text-decoration: none;
  }

  &.active {
    background: transparent;
    font-weight: 500;
    color: $color-accent;
    border-left: 2px solid $color-accent;
    padding-left: 14px;
  }
}

.admin-content {
  flex: 1;
  padding: 32px 48px;
  min-width: 0;
}

@media (max-width: 768px) {
  .admin-sidebar {
    width: 60px;
    padding: 16px 8px;
  }

  .nav-item span {
    display: none;
  }

  .admin-content {
    padding: 24px 16px;
  }
}
</style>
