import { apiClient } from './client'

export type AIToolStatus = 'draft' | 'published' | 'archived'

export interface AITool {
  id: number
  name: string
  category: string
  description: string
  url: string
  status: AIToolStatus
  sort_order: number
  created_at?: string
  updated_at?: string
}

export interface AIToolPayload {
  name: string
  category: string
  description?: string
  url: string
  status: AIToolStatus
  sort_order?: number
}

export const aiToolsAPI = {
  list() {
    return apiClient.get<AITool[]>('/tools')
  },
}

export const adminAIToolsAPI = {
  list() {
    return apiClient.get<AITool[]>('/admin/tools')
  },
  create(data: AIToolPayload) {
    return apiClient.post<AITool>('/admin/tools', data)
  },
  update(id: number, data: AIToolPayload) {
    return apiClient.put<AITool>(`/admin/tools/${id}`, data)
  },
  delete(id: number) {
    return apiClient.delete(`/admin/tools/${id}`)
  },
}
