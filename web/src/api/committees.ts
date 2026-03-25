import apiClient from './client'
import type { APIResponse } from '../types'

export interface Committee {
  id: string
  name: string
  description: string
  chair_id?: string
  meeting_schedule?: string
  created_at: string
}

export async function getCommittees(): Promise<Committee[]> {
  const { data } = await apiClient.get<APIResponse<Committee[]>>('/committees')
  return data.data ?? []
}

export async function getCommittee(id: string): Promise<Committee> {
  const { data } = await apiClient.get<APIResponse<Committee>>(`/committees/${id}`)
  return data.data
}

export async function createCommittee(input: Partial<Committee>): Promise<Committee> {
  const { data } = await apiClient.post<APIResponse<Committee>>('/committees', input)
  return data.data
}

export async function joinCommittee(id: string): Promise<void> {
  await apiClient.post(`/committees/${id}/members`, { role: 'member' })
}

export async function leaveCommittee(committeeId: string, memberId: string): Promise<void> {
  await apiClient.delete(`/committees/${committeeId}/members/${memberId}`)
}
