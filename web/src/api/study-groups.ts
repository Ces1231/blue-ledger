import apiClient from './client'
import type { APIResponse } from '../types'

export interface StudyGroup {
  id: string
  name: string
  subject: string
  description?: string
  created_by: string
  member_count: number
  created_at: string
}

export async function getStudyGroups(): Promise<StudyGroup[]> {
  const { data } = await apiClient.get<APIResponse<StudyGroup[]>>('/study-groups')
  return data.data ?? []
}

export async function joinStudyGroup(id: string): Promise<void> {
  await apiClient.post(`/study-groups/${id}/join`)
}

export async function createStudyGroup(input: Partial<StudyGroup>): Promise<StudyGroup> {
  const { data } = await apiClient.post<APIResponse<StudyGroup>>('/study-groups', input)
  return data.data
}
