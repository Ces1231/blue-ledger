import apiClient from './client'
import type { APIResponse } from '../types'

export interface StudyGroup {
  id: string
  chapter_id: string
  topic: string
  host_id: string
  host_name?: string
  date: string          // "YYYY-MM-DD"
  location?: string
  member_ids: string[]
  xp_reward: number
  created_at: string
}

export interface CreateStudyGroupPayload {
  topic: string
  date: string
  location?: string
  xp_reward?: number
}

export async function getStudyGroups(): Promise<StudyGroup[]> {
  const { data } = await apiClient.get<APIResponse<StudyGroup[]>>('/study-groups')
  return data.data ?? []
}

export async function createStudyGroup(input: CreateStudyGroupPayload): Promise<StudyGroup> {
  const { data } = await apiClient.post<APIResponse<StudyGroup>>('/study-groups', input)
  return data.data
}

export async function joinStudyGroup(id: string): Promise<void> {
  await apiClient.post(`/study-groups/${id}/join`)
}

export async function deleteStudyGroup(id: string): Promise<void> {
  await apiClient.delete(`/study-groups/${id}`)
}
