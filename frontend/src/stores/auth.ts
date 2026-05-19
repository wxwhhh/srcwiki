import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types'
import { getMe, logout as apiLogout } from '@/api/bff'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const loading = ref(false)

  const isLoggedIn = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isEditor = computed(() => user.value?.role === 'editor' || user.value?.role === 'admin')

  async function fetchUser() {
    loading.value = true
    try {
      const res = await getMe()
      if (res.code === 0) {
        user.value = res.data
      }
    } catch {
      user.value = null
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      await apiLogout()
    } finally {
      user.value = null
      window.location.href = '/admin/login'
    }
  }

  function setUser(u: User) {
    user.value = u
  }

  return { user, loading, isLoggedIn, isAdmin, isEditor, fetchUser, logout, setUser }
})
