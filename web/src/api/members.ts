import apiClient from './client'
import type { Member, EngagementLog, PaginatedResponse, APIResponse } from '../types'

export async function getMembers(params?: { page?: number; per_page?: number }): Promise<PaginatedResponse<Member>> {
  const { data } = await apiClient.get<PaginatedResponse<Member>>('/members', { params })
  return data
}

export async function getMember(id: string): Promise<Member> {
  const { data } = await apiClient.get<APIResponse<Member>>(`/members/${id}`)
  return data.data
}

export interface UpdateMemberPayload {
  inducted_year?: number
  employer?: string
  job_title?: string
  city?: string
  linkedin_url?: string
  avatar_bg?: string
  avatar_fg?: string
  // Admin only
  role?: string
  status?: string
  dues_status?: string
}

export async function updateMember(id: string, payload: UpdateMemberPayload): Promise<Member> {
  const { data } = await apiClient.put<APIResponse<Member>>(`/members/${id}`, payload)
  return data.data
}

export async function deleteMember(id: string): Promise<void> {
  await apiClient.delete(`/members/${id}`)
}

export async function getMemberXPHistory(
  id: string,
  params?: { page?: number; per_page?: number }
): Promise<PaginatedResponse<EngagementLog>> {
  const { data } = await apiClient.get<PaginatedResponse<EngagementLog>>(
    `/members/${id}/xp-history`,
    { params }
  )
  return data
}
