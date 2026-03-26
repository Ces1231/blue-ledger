import apiClient from './client'
import type { APIResponse } from '../types'

export interface Milestone {
  id: string
  chapter_id: string
  member_id: string
  member_name?: string
  type: string
  title: string
  date?: string
  description?: string
  created_at: string
}

export interface CreateMilestonePayload {
  member_id: string
  type: string
  title: string
  date?: string
  description?: string
}

export async function getMilestones(memberID?: string): Promise<Milestone[]> {
  const params = memberID ? { member_id: memberID } : undefined
  const { data } = await apiClient.get<APIResponse<Milestone[]>>('/milestones', { params })
  return data.data ?? []
}

export async function createMilestone(input: CreateMilestonePayload): Promise<Milestone> {
  const { data } = await apiClient.post<APIResponse<Milestone>>('/milestones', input)
  return data.data
}

export async function deleteMilestone(id: string): Promise<void> {
  await apiClient.delete(`/milestones/${id}`)
}
