import apiClient from './client'
import type { APIResponse } from '../types'

export interface Minutes {
  id: string
  chapter_id: string
  title: string
  meeting_date: string
  body: string
  created_by: string
  is_finalized: boolean
  finalized_at?: string
  finalized_by?: string
  created_at: string
  updated_at: string
  author_first_name?: string
  author_last_name?: string
}

export interface MinutesListResponse {
  data: Minutes[]
  meta: { total: number }
}

export async function getMinutes(params?: {
  page?: number
  per_page?: number
}): Promise<MinutesListResponse> {
  const { data } = await apiClient.get<MinutesListResponse>('/minutes', { params })
  return data
}

export async function getMinutesById(id: string): Promise<Minutes> {
  const { data } = await apiClient.get<APIResponse<Minutes>>(`/minutes/${id}`)
  return data.data
}

export interface CreateMinutesPayload {
  title: string
  meeting_date: string
  body: string
}

export async function createMinutes(payload: CreateMinutesPayload): Promise<Minutes> {
  const { data } = await apiClient.post<APIResponse<Minutes>>('/minutes', payload)
  return data.data
}

export async function updateMinutes(
  id: string,
  payload: Partial<CreateMinutesPayload>
): Promise<Minutes> {
  const { data } = await apiClient.put<APIResponse<Minutes>>(`/minutes/${id}`, payload)
  return data.data
}

export async function finalizeMinutes(id: string): Promise<Minutes> {
  const { data } = await apiClient.put<APIResponse<Minutes>>(`/minutes/${id}/finalize`, {})
  return data.data
}
