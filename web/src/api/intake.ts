import apiClient from './client'
import type { APIResponse } from '../types'

export interface Prospect {
  id: string
  chapter_id: string
  first_name: string
  last_name: string
  email?: string
  phone?: string
  university?: string
  grad_year?: number
  stage: string
  notes?: string
  added_by: string
  created_at: string
  updated_at: string
  archived_at?: string
}

export async function getProspects(): Promise<Prospect[]> {
  const { data } = await apiClient.get<APIResponse<Prospect[]>>('/intake')
  return data.data
}

export interface CreateProspectPayload {
  first_name: string
  last_name: string
  email?: string
  phone?: string
  university?: string
  grad_year?: number
  notes?: string
}

export async function createProspect(payload: CreateProspectPayload): Promise<Prospect> {
  const { data } = await apiClient.post<APIResponse<Prospect>>('/intake', payload)
  return data.data
}

export async function updateProspect(id: string, payload: Partial<CreateProspectPayload>): Promise<Prospect> {
  const { data } = await apiClient.put<APIResponse<Prospect>>(`/intake/${id}`, payload)
  return data.data
}

export async function updateProspectStage(id: string, stage: string): Promise<Prospect> {
  const { data } = await apiClient.put<APIResponse<Prospect>>(`/intake/${id}/stage`, { stage })
  return data.data
}

export async function archiveProspect(id: string): Promise<void> {
  await apiClient.delete(`/intake/${id}`)
}
