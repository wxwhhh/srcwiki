// TypeScript 类型定义

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export interface User {
  id: number
  username: string
  role: 'admin' | 'editor' | 'reader'
  status: 'active' | 'disabled'
  created_at: string
  updated_at: string
}

export interface InviteCode {
  id: number
  code: string
  role: 'editor' | 'reader'
  max_uses: number
  use_count: number
  expires_at: string | null
  created_by: number
  created_at: string
}

export interface Category {
  id: number
  parent_id: number | null
  name: string
  sort_order: number
  doc_count?: number
  created_at: string
  updated_at: string
  children?: Category[]
}

export interface Document {
  id: number
  category_id: number | null
  title: string
  content: string
  author_id: number
  is_published: boolean
  created_at: string
  updated_at: string
  author_name?: string
  category_name?: string
}

export interface DocumentVersion {
  id: number
  document_id: number
  title: string
  content: string
  editor_id: number
  created_at: string
  editor_name?: string
}

export interface AuditLog {
  id: number
  user_id: number
  username: string
  action: string
  target_type: string
  target_id: number
  detail: string
  ip: string
  user_agent: string
  created_at: string
}

export interface TreeNode {
  id: number
  parent_id: number | null
  name: string
  sort_order: number
  children?: TreeNode[]
  docs?: TreeDoc[]
}

export interface TreeDoc {
  id: number
  title: string
  is_published: boolean
}

export interface SearchHit {
  id: number
  title: string
  content_snippet: string
  category_id: number | null
  category_name: string
  updated_at: string
}

export interface PaginatedData<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface Credit {
  id: number
  name: string
  url: string
  description: string
  icon_url: string
  license: string
  stars: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface RecentDocItem {
  id: number
  title: string
  category_name: string
  updated_at: string
}

export interface HomeStats {
  total_docs: number
  total_cats: number
  recent_docs: RecentDocItem[]
}
