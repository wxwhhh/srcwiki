import axios from 'axios'
import type { ApiResponse, User, InviteCode, Category, Document, DocumentVersion, AuditLog, PaginatedData, Credit } from '@/types'

const api = axios.create({
  baseURL: '',
  withCredentials: true,
  timeout: 15000,
})

// 请求拦截器：自动添加 Authorization header
api.interceptors.request.use((config) => {
  const token = document.cookie.split('; ').find(c => c.startsWith('token='))?.split('=')[1]
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      window.location.href = '/admin/login'
    }
    return Promise.reject(err)
  }
)

// ========== Users ==========

export async function listUsers(page = 1, page_size = 20): Promise<ApiResponse<PaginatedData<User>>> {
  const res = await api.get('/api/admin/users', { params: { page, page_size } })
  return res.data
}

export async function updateUserRole(id: number, role: string): Promise<ApiResponse> {
  const res = await api.put(`/api/admin/users/${id}/role`, { role })
  return res.data
}

export async function updateUserStatus(id: number, status: string): Promise<ApiResponse> {
  const res = await api.put(`/api/admin/users/${id}/status`, { status })
  return res.data
}

export async function deleteUser(id: number): Promise<ApiResponse> {
  const res = await api.delete(`/api/admin/users/${id}`)
  return res.data
}

// ========== Invite Codes ==========

export async function listInvites(page = 1, page_size = 20): Promise<ApiResponse<PaginatedData<InviteCode>>> {
  const res = await api.get('/api/admin/invites', { params: { page, page_size } })
  return res.data
}

export async function createInvite(data: { role: string; max_uses: number; expires_at?: string }): Promise<ApiResponse<InviteCode>> {
  const res = await api.post('/api/admin/invites', data)
  return res.data
}

export async function batchCreateInvites(data: { count: number; role: string; expires_in_hours?: number; max_uses?: number }): Promise<ApiResponse<{ codes: InviteCode[]; count: number }>> {
  const res = await api.post('/api/admin/invites/batch', data)
  return res.data
}

export async function deleteInvite(id: number): Promise<ApiResponse> {
  const res = await api.delete(`/api/admin/invites/${id}`)
  return res.data
}

// ========== Categories ==========

export async function listCategories(): Promise<ApiResponse<Category[]>> {
  const res = await api.get('/api/admin/categories')
  return res.data
}

export async function createCategory(data: { parent_id?: number | null; name: string; sort_order?: number }): Promise<ApiResponse<Category>> {
  const res = await api.post('/api/admin/categories', data)
  return res.data
}

export async function updateCategory(id: number, data: { name: string; sort_order?: number; parent_id?: number | null }): Promise<ApiResponse> {
  const res = await api.put(`/api/admin/categories/${id}`, data)
  return res.data
}

export async function deleteCategory(id: number): Promise<ApiResponse> {
  const res = await api.delete(`/api/admin/categories/${id}`)
  return res.data
}

export async function batchDeleteCategories(ids: number[]): Promise<ApiResponse<{ deleted: number }>> {
  const res = await api.post('/api/admin/categories/batch-delete', { ids })
  return res.data
}

export async function sortCategories(items: { id: number; sort_order: number }[]): Promise<ApiResponse> {
  const res = await api.put('/api/admin/categories/sort', { items })
  return res.data
}

export async function cascadeDeleteCategories(ids: number[]): Promise<ApiResponse<{ deleted_categories: number; deleted_documents: number }>> {
  const res = await api.post('/api/admin/categories/cascade-delete', { ids })
  return res.data
}

// ========== Documents ==========

export async function listDocuments(params: { page?: number; size?: number; category_id?: number; status?: string }): Promise<ApiResponse<PaginatedData<Document>>> {
  const res = await api.get('/api/admin/documents', { params })
  return res.data
}

export async function getAdminDocument(id: number): Promise<ApiResponse<Document>> {
  const res = await api.get(`/api/admin/documents/${id}`)
  return res.data
}

export async function createDocument(data: { title: string; content: string; category_id?: number }): Promise<ApiResponse<Document>> {
  const res = await api.post('/api/admin/documents', data)
  return res.data
}

export async function updateDocument(id: number, data: { title: string; content: string; category_id?: number }): Promise<ApiResponse> {
  const res = await api.put(`/api/admin/documents/${id}`, data)
  return res.data
}

export async function deleteDocument(id: number): Promise<ApiResponse> {
  const res = await api.delete(`/api/admin/documents/${id}`)
  return res.data
}

export async function batchDeleteDocuments(ids: number[]): Promise<ApiResponse<{ deleted: number }>> {
  const res = await api.post('/api/admin/documents/batch-delete', { ids })
  return res.data
}

export async function publishDocument(id: number, is_published: boolean): Promise<ApiResponse> {
  const res = await api.put(`/api/admin/documents/${id}/publish`, { is_published })
  return res.data
}

export async function getDocumentVersions(id: number): Promise<ApiResponse<DocumentVersion[]>> {
  const res = await api.get(`/api/admin/documents/${id}/versions`)
  return res.data
}

export async function rollbackDocument(id: number, vid: number): Promise<ApiResponse> {
  const res = await api.post(`/api/admin/documents/${id}/rollback/${vid}`)
  return res.data
}

// ========== Audit Log ==========

export async function listAuditLogs(params: { page?: number; page_size?: number; user_id?: number; action?: string; start_date?: string; end_date?: string }): Promise<ApiResponse<PaginatedData<AuditLog>>> {
  const res = await api.get('/api/admin/audit-log', { params })
  return res.data
}

export interface LoginStats {
  total_logins: number
  today_logins: number
  active_users: number
  logins_by_user: { username: string; count: number; last_ip: string; last_login: string }[]
  recent_logins: { id: number; username: string; ip: string; user_agent: string; created_at: string }[]
}

export async function getLoginStats(): Promise<ApiResponse<LoginStats>> {
  const res = await api.get('/api/admin/audit/login-stats')
  return res.data
}

// ========== File Upload ==========

export async function uploadFile(file: File): Promise<ApiResponse<{ url: string }>> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await api.post('/api/admin/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return res.data
}

// ========== Import ========== (legacy sync)

export async function importZip(file: File): Promise<ApiResponse<{
  categories_created: number
  categories_skipped: number
  docs_created: number
  docs_updated: number
  docs_skipped: number
  errors: string[]
}>> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await api.post('/api/admin/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 300000,
  })
  return res.data
}

// ========== Import Tasks (async) ==========

export interface ImportTask {
  id: number
  type: 'zip' | 'github'
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  source: string
  progress: number
  total_docs: number
  imported_docs: number
  updated_docs: number
  skipped_docs: number
  error_count: number
  errors: string
  result: string
  created_at: string
  started_at: string | null
  finished_at: string | null
}

export async function createZipImportTask(file: File): Promise<ApiResponse<{ task_id: number; message: string }>> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await api.post('/api/admin/import/async', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 60000,
  })
  return res.data
}

export async function createBatchZipImportTask(files: File[]): Promise<ApiResponse<{ task_ids: number[]; count: number; message: string }>> {
  const formData = new FormData()
  files.forEach(f => formData.append('files', f))
  const res = await api.post('/api/admin/import/batch', formData, {
    timeout: 120000,
  })
  return res.data
}

export async function createGithubImportTask(data: { url: string; branch?: string; skip_root?: boolean }): Promise<ApiResponse<{ task_id: number; message: string }>> {
  const res = await api.post('/api/admin/import/github/async', data)
  return res.data
}

export async function listImportTasks(page = 1, page_size = 20): Promise<ApiResponse<PaginatedData<ImportTask>>> {
  const res = await api.get('/api/admin/import/tasks', { params: { page, page_size } })
  return res.data
}

export async function getImportTask(id: number): Promise<ApiResponse<ImportTask>> {
  const res = await api.get(`/api/admin/import/tasks/${id}`)
  return res.data
}

export async function retryImportTask(id: number): Promise<ApiResponse> {
  const res = await api.post(`/api/admin/import/tasks/${id}/retry`)
  return res.data
}

export async function deleteImportTask(id: number): Promise<ApiResponse> {
  const res = await api.delete(`/api/admin/import/tasks/${id}`)
  return res.data
}

export async function cancelImportTask(id: number): Promise<ApiResponse> {
  const res = await api.post(`/api/admin/import/tasks/${id}/cancel`)
  return res.data
}

// ========== Credits ==========

export async function getCredits(): Promise<ApiResponse<Credit[]>> {
  const res = await api.get('/bff/credits')
  return res.data
}

export async function getAdminCredits(): Promise<ApiResponse<Credit[]>> {
  const res = await api.get('/api/admin/credits')
  return res.data
}

export async function createCredit(data: { name: string; url: string; description?: string; icon_url?: string; license?: string; stars?: string; sort_order?: number }): Promise<ApiResponse<Credit>> {
  const res = await api.post('/api/admin/credits', data)
  return res.data
}

export async function updateCredit(id: number, data: { name: string; url: string; description?: string; icon_url?: string; license?: string; stars?: string; sort_order?: number }): Promise<ApiResponse<Credit>> {
  const res = await api.put(`/api/admin/credits/${id}`, data)
  return res.data
}

export async function deleteCredit(id: number): Promise<ApiResponse> {
  const res = await api.delete(`/api/admin/credits/${id}`)
  return res.data
}

// ========== Settings ==========

export async function getAdminSettings(): Promise<ApiResponse<{ register_mode: string }>> {
  const res = await api.get('/api/admin/settings')
  return res.data
}

export async function updateAdminSettings(data: { register_mode: string }): Promise<ApiResponse<{ register_mode: string }>> {
  const res = await api.put('/api/admin/settings', data)
  return res.data
}

// ========== Dashboard Stats ==========

export async function getDashboardStats(): Promise<ApiResponse<{ user_count: number; doc_count: number; category_count: number; recent_logs: AuditLog[] }>> {
  // 通过并行请求获取统计数据
  const [usersRes, docsRes, catsRes, logsRes] = await Promise.all([
    api.get('/api/admin/users', { params: { page: 1, page_size: 1 } }),
    api.get('/api/admin/documents', { params: { page: 1, size: 1 } }),
    api.get('/api/admin/categories'),
    api.get('/api/admin/audit-log', { params: { page: 1, page_size: 10 } }),
  ])
  return {
    code: 0,
    message: 'success',
    data: {
      user_count: usersRes.data?.data?.total || 0,
      doc_count: docsRes.data?.data?.total || 0,
      category_count: catsRes.data?.data?.length || 0,
      recent_logs: logsRes.data?.data?.list || [],
    },
  }
}
