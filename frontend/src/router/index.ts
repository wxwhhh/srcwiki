import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  // 首页重定向
  {
    path: '/',
    redirect: '/admin/',
  },

  // 公开页面
  {
    path: '/admin/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { guest: true },
  },
  {
    path: '/admin/register',
    name: 'Register',
    component: () => import('@/views/Register.vue'),
    meta: { guest: true },
  },

  // 前台文档查看（独立路由，需登录）
  {
    path: '/doc/:id',
    name: 'DocView',
    component: () => import('@/views/Document.vue'),
    props: true,
    meta: { requiresAuth: true },
  },

  // 前台页面（需登录）
  {
    path: '/admin/',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('@/views/Home.vue'),
      },
      {
        path: 'doc/:id',
        name: 'Document',
        component: () => import('@/views/Document.vue'),
        props: true,
      },
      {
        path: 'search',
        name: 'Search',
        component: () => import('@/views/Search.vue'),
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/Profile.vue'),
      },
    ],
  },

  // 管理后台
  {
    path: '/admin/dashboard',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAuth: true, requiresAdmin: true },
    children: [
      {
        path: '',
        name: 'AdminDashboard',
        component: () => import('@/views/admin/Dashboard.vue'),
      },
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/Users.vue'),
      },
      {
        path: 'invites',
        name: 'AdminInvites',
        component: () => import('@/views/admin/InviteCodes.vue'),
      },
      {
        path: 'content',
        name: 'AdminContent',
        component: () => import('@/views/admin/DocumentsManager.vue'),
      },
      {
        path: 'import',
        name: 'AdminImport',
        component: () => import('@/views/admin/Import.vue'),
      },
      {
        path: 'tasks',
        name: 'AdminImportTasks',
        component: () => import('@/views/admin/ImportTasks.vue'),
      },
      {
        path: 'audit',
        name: 'AdminAuditLog',
        component: () => import('@/views/admin/AuditLog.vue'),
      },
      {
        path: 'credits',
        name: 'AdminCredits',
        component: () => import('@/views/admin/Credits.vue'),
      },
      {
        path: 'settings',
        name: 'AdminSettings',
        component: () => import('@/views/admin/Settings.vue'),
      },
      // 旧路由重定向
      {
        path: 'categories',
        redirect: '/admin/dashboard/content',
      },
      {
        path: 'documents',
        redirect: '/admin/dashboard/content',
      },
      {
        path: 'documents/:id',
        redirect: '/admin/dashboard/content',
      },
    ],
  },

  // 404
  {
    path: '/:pathMatch(.*)*',
    redirect: '/admin/',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach(async (to, _from, next) => {
  const auth = useAuthStore()

  // 如果用户信息未加载，先获取
  if (!auth.user && !to.meta.guest) {
    await auth.fetchUser()
  }

  // 需要登录
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
    return
  }

  // 需要管理员
  if (to.meta.requiresAdmin && !auth.isAdmin) {
    next({ name: 'Home' })
    return
  }

  // 已登录用户访问登录页，重定向首页
  if (to.meta.guest && auth.isLoggedIn) {
    next({ name: 'Home' })
    return
  }

  next()
})

export default router
