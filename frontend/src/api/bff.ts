import axios from 'axios'
import type { ApiResponse, TreeNode, Document, SearchHit, PaginatedData, User, Credit, HomeStats } from '@/types'

const api = axios.create({
  baseURL: '',
  withCredentials: true,
  timeout: 15000,
})

// 统一错误处理
api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      window.location.href = '/admin/login'
    }
    return Promise.reject(err)
  }
)

// ========== Settings ==========

export async function getPublicSettings(): Promise<ApiResponse<{ register_mode: string }>> {
  const res = await api.get('/bff/settings')
  return res.data
}

// ========== Auth ==========

export async function login(username: string, password: string, captcha_id: string, captcha_code: string): Promise<ApiResponse> {
  const res = await api.post('/bff/auth/login', { username, password, captcha_id, captcha_code })
  return res.data
}

export async function getCaptcha(): Promise<ApiResponse<{ captcha_id: string; captcha_img: string }>> {
  const res = await api.get('/bff/captcha')
  return res.data
}

export async function register(username: string, password: string, invite_code: string, captcha_id: string, captcha_code: string): Promise<ApiResponse> {
  const res = await api.post('/bff/auth/register', { username, password, invite_code, captcha_id, captcha_code })
  return res.data
}

export async function logout(): Promise<ApiResponse> {
  const res = await api.post('/bff/auth/logout')
  return res.data
}

export async function getMe(): Promise<ApiResponse<User>> {
  const res = await api.get('/bff/auth/me')
  return res.data
}

// ========== Tree ==========

export async function getTree(): Promise<ApiResponse<TreeNode[]>> {
  const res = await api.get('/bff/tree')
  return res.data
}

// ========== Documents ==========

export async function getDocument(id: number): Promise<ApiResponse<Document>> {
  const res = await api.get(`/bff/docs/${id}`)
  return res.data
}

export async function listDocsByCategory(categoryId: number): Promise<ApiResponse<Document[]>> {
  const res = await api.get(`/bff/docs/category/${categoryId}`)
  return res.data
}

// ========== Search ==========

export async function search(q: string): Promise<ApiResponse<PaginatedData<SearchHit>>> {
  const res = await api.get('/bff/search', { params: { q } })
  return res.data
}

// ========== User Profile ==========

export async function updateProfile(username: string): Promise<ApiResponse> {
  const res = await api.put('/bff/user/profile', { username })
  return res.data
}

export async function changePassword(old_password: string, new_password: string): Promise<ApiResponse> {
  const res = await api.put('/bff/user/password', { old_password, new_password })
  return res.data
}

// ========== Credits ==========

export async function getCredits(): Promise<ApiResponse<Credit[]>> {
  const res = await api.get('/bff/credits')
  return res.data
}

// ========== Home Stats ==========

export async function getHomeStats(): Promise<ApiResponse<HomeStats>> {
  const res = await api.get(`/bff/home-stats`)
  return res.data
}
