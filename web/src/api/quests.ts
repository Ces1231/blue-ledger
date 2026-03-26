import apiClient from './client'
import type { APIResponse } from '../types'

export interface Quest {
  id: string
  name: string
  description: string
  requirements: Record<string, unknown>
  xp_reward: number
  badge_id: string | null
  active: boolean
  created_at: string
}

export interface QuestProgress {
  id?: string
  member_id: string
  quest_id: string
  status: 'not_started' | 'in_progress' | 'completed'
  completed_at?: string
}

export async function getQuests(): Promise<Quest[]> {
  const { data } = await apiClient.get<APIResponse<Quest[]>>('/quests')
  return data.data ?? []
}

export async function getQuestProgress(questId: string): Promise<QuestProgress> {
  const { data } = await apiClient.get<APIResponse<QuestProgress>>(`/quests/${questId}/progress`)
  return data.data
}
