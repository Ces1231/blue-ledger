import apiClient from './client'
import type { APIResponse } from '../types'

export interface ScholarshipApplication {
  id: string
  chapter_id: string
  member_id: string
  title: string
  amount_cents: number
  provider?: string
  deadline?: string
  status: string
  notes?: string
  applied_at?: string
  awarded_at?: string
  created_at: string
  updated_at: string
  member_first_name?: string
  member_last_name?: string
}

export interface ScholarshipListResponse {
  data: ScholarshipApplication[]
  meta: { total: number }
}

export async function getScholarships(params?: {
  page?: number
  per_page?: number
}): Promise<ScholarshipListResponse> {
  const { data } = await apiClient.get<ScholarshipListResponse>('/scholarships', { params })
  return data
}

export async function getScholarship(id: string): Promise<ScholarshipApplication> {
  const { data } = await apiClient.get<APIResponse<ScholarshipApplication>>(`/scholarships/${id}`)
  return data.data
}

export interface CreateScholarshipPayload {
  title: string
  amount_cents?: number
  provider?: string
  deadline?: string
  notes?: string
}

export async function createScholarship(payload: CreateScholarshipPayload): Promise<ScholarshipApplication> {
  const { data } = await apiClient.post<APIResponse<ScholarshipApplication>>('/scholarships', payload)
  return data.data
}

export async function updateScholarship(
  id: string,
  payload: Partial<CreateScholarshipPayload> & { status?: string }
): Promise<ScholarshipApplication> {
  const { data } = await apiClient.put<APIResponse<ScholarshipApplication>>(`/scholarships/${id}`, payload)
  return data.data
}
