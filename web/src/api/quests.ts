import apiClient from './client'
import type { APIResponse } from '../types'

export interface Quest {
  id: string
  chapter_id: string
  title: string
  description?: string
  xp_reward: number
  badge_reward_id?: string | null
  steps: Record<string, unknown>[]
  is_active: boolean
  created_at: string
}

export interface QuestProgress {
  id?: string
  chapter_id?: string
  member_id: string
  quest_id: string
  progress: Record<string, number>
  completed_at?: string
}

export async function getQuests(): Promise<Quest[]> {
  const { data } = await apiClient.get<APIResponse<Quest[]>>('/quests')
  return data.data ?? []
}

export async function getQuest(questId: string): Promise<Quest> {
  const { data } = await apiClient.get<APIResponse<Quest>>(`/quests/${questId}`)
  return data.data
}

export async function getQuestProgress(questId: string): Promise<QuestProgress> {
  const { data } = await apiClient.get<APIResponse<QuestProgress>>(`/quests/${questId}/progress`)
  return data.data
}
