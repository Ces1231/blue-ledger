import apiClient from './client'
import type { APIResponse } from '../types'

export interface Badge {
  id: string
  name: string
  description: string
  criteria: Record<string, unknown>
  xp_bonus: number
  image_url?: string
  created_at: string
}

export interface MemberBadge {
  id: string
  member_id: string
  badge_id: string
  earned_at: string
  awarded_by?: string
}

export async function getBadges(): Promise<Badge[]> {
  const { data } = await apiClient.get<APIResponse<Badge[]>>('/badges')
  return data.data ?? []
}

export async function createBadge(input: Partial<Badge>): Promise<Badge> {
  const { data } = await apiClient.post<APIResponse<Badge>>('/badges', input)
  return data.data
}

export async function updateBadge(id: string, input: Partial<Badge>): Promise<Badge> {
  const { data } = await apiClient.put<APIResponse<Badge>>(`/badges/${id}`, input)
  return data.data
}

export async function deleteBadge(id: string): Promise<void> {
  await apiClient.delete(`/badges/${id}`)
}

export async function awardBadge(badgeId: string, memberId: string): Promise<MemberBadge> {
  const { data } = await apiClient.post<APIResponse<MemberBadge>>(`/badges/${badgeId}/award`, { member_id: memberId })
  return data.data
}
