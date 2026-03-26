import apiClient from './client'
import type { APIResponse } from '../types'

export interface Mentor {
  id: string
  chapter_id: string
  member_id: string
  bio?: string
  specialties: string[]
  is_active: boolean
  created_at: string
  first_name?: string
  last_name?: string
  avatar_url?: string
  role?: string
}

export interface MentorshipMatch {
  id: string
  chapter_id: string
  mentor_id: string
  mentee_id: string
  status: 'pending' | 'active' | 'ended'
  started_at?: string
  ended_at?: string
  created_at: string
  mentor_first_name?: string
  mentor_last_name?: string
  mentee_first_name?: string
  mentee_last_name?: string
}

export async function getMentors(): Promise<Mentor[]> {
  const { data } = await apiClient.get<APIResponse<Mentor[]>>('/mentorship')
  return data.data
}

export interface BecomeMentorPayload {
  bio?: string
  specialties?: string[]
}

export async function becomeMentor(payload: BecomeMentorPayload): Promise<Mentor> {
  const { data } = await apiClient.post<APIResponse<Mentor>>('/mentorship/become-mentor', payload)
  return data.data
}

export async function requestMentorshipMatch(mentorId: string): Promise<MentorshipMatch> {
  const { data } = await apiClient.post<APIResponse<MentorshipMatch>>(`/mentorship/${mentorId}/request`, {})
  return data.data
}

export async function getMyMatch(): Promise<MentorshipMatch | null> {
  const { data } = await apiClient.get<APIResponse<MentorshipMatch | null>>('/mentorship/my-match')
  return data.data
}
