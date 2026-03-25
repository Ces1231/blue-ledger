import apiClient from './client'
import type { APIResponse } from '../types'

export interface Vote {
  id: string
  chapter_id: string
  title: string
  description?: string
  options: string[]
  is_open: boolean
  created_by: string
  closes_at?: string
  created_at: string
  my_response?: string
}

export interface VoteResult {
  vote_id: string
  total: number
  options: Record<string, number>
}

export async function getVotes(): Promise<Vote[]> {
  const { data } = await apiClient.get<APIResponse<Vote[]>>('/votes')
  return data.data
}

export interface CreateVotePayload {
  title: string
  description?: string
  options: string[]
  closes_at?: string
}

export async function createVote(payload: CreateVotePayload): Promise<Vote> {
  const { data } = await apiClient.post<APIResponse<Vote>>('/votes', payload)
  return data.data
}

export async function respondToVote(voteId: string, choice: string): Promise<void> {
  await apiClient.post(`/votes/${voteId}/respond`, { choice })
}

export async function getVoteResults(voteId: string): Promise<VoteResult> {
  const { data } = await apiClient.get<APIResponse<VoteResult>>(`/votes/${voteId}/results`)
  return data.data
}
