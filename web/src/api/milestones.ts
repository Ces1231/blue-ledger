import apiClient from './client'
import type { APIResponse } from '../types'

export interface Milestone {
  id: string
  member_id: string
  type: string
  title: string
  description?: string
  date?: string
  created_at: string
}

export async function getMilestones(): Promise<Milestone[]> {
  const { data } = await apiClient.get<APIResponse<Milestone[]>>('/milestones')
  return data.data ?? []
}
