import apiClient from './client'
import type { APIResponse } from '../types'

export interface Goal {
  id: string
  chapter_id: string
  title: string
  description?: string
  category: string
  target_value: number
  current_value: number
  xp_reward: number
  is_active: boolean
  due_date?: string
  created_at: string
  updated_at: string
}

export async function getGoals(): Promise<Goal[]> {
  const { data } = await apiClient.get<APIResponse<Goal[]>>('/goals')
  return data.data
}

export interface CreateGoalPayload {
  title: string
  description?: string
  category: string
  target_value: number
  xp_reward?: number
  due_date?: string
}

export async function createGoal(payload: CreateGoalPayload): Promise<Goal> {
  const { data } = await apiClient.post<APIResponse<Goal>>('/goals', payload)
  return data.data
}

export async function updateGoal(id: string, payload: Partial<CreateGoalPayload> & { is_active?: boolean }): Promise<Goal> {
  const { data } = await apiClient.put<APIResponse<Goal>>(`/goals/${id}`, payload)
  return data.data
}

export async function updateGoalProgress(id: string, increment: number): Promise<Goal> {
  const { data } = await apiClient.put<APIResponse<Goal>>(`/goals/${id}/progress`, { increment })
  return data.data
}
