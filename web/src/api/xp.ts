import apiClient from './client'
import type { LeaderboardEntry, EngagementLog, PaginatedResponse, APIResponse } from '../types'

export async function getLeaderboard(mode?: 'alltime' | 'semester'): Promise<LeaderboardEntry[]> {
  const { data } = await apiClient.get<APIResponse<LeaderboardEntry[]>>('/leaderboard', {
    params: { mode: mode ?? 'alltime' },
  })
  return data.data
}

export interface AwardXPPayload {
  member_id: string
  xp_amount: number
  activity: string
  note?: string
  semester?: string
}

export async function awardXP(payload: AwardXPPayload): Promise<void> {
  await apiClient.post('/xp/award', payload)
}

export async function recalculateXP(): Promise<{ members_updated: number }> {
  const { data } = await apiClient.post<APIResponse<{ members_updated: number }>>('/xp/recalculate')
  return data.data
}

export async function getEngagementLog(params?: {
  page?: number
  per_page?: number
}): Promise<PaginatedResponse<EngagementLog>> {
  const { data } = await apiClient.get<PaginatedResponse<EngagementLog>>('/engagement-log', { params })
  return data
}
