import apiClient from './client'
import type { APIResponse } from '../types'

export interface Resource {
  id: string
  title: string
  description?: string
  url: string
  category?: string
  uploaded_by: string
  created_at: string
}

export async function getResources(): Promise<Resource[]> {
  const { data } = await apiClient.get<APIResponse<Resource[]>>('/resources')
  return data.data ?? []
}

export async function createResource(input: Partial<Resource>): Promise<Resource> {
  const { data } = await apiClient.post<APIResponse<Resource>>('/resources', input)
  return data.data
}

export async function deleteResource(id: string): Promise<void> {
  await apiClient.delete(`/resources/${id}`)
}
