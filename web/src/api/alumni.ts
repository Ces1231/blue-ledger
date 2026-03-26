import apiClient from './client'
import type { APIResponse } from '../types'

export interface Alumnus {
  id: string
  name: string
  graduation_year?: number
  employer?: string
  title?: string
  city?: string
  linkedin_url?: string
  created_at: string
}

export async function getAlumni(): Promise<Alumnus[]> {
  const { data } = await apiClient.get<APIResponse<Alumnus[]>>('/alumni')
  return data.data ?? []
}

export async function createAlumnus(input: Partial<Alumnus>): Promise<Alumnus> {
  const { data } = await apiClient.post<APIResponse<Alumnus>>('/alumni', input)
  return data.data
}
