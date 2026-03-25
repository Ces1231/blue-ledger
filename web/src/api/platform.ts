import apiClient from './client'
import type { APIResponse } from '../types'

export interface ChapterSummary {
  id: string
  name: string
  greek_letters: string
  city: string
  state_code: string
  university?: string
  member_count: number
  subscription_status: string
  plan_tier: string
  trial_ends_at?: string
  created_at: string
}

export interface SysUser {
  id: string
  email: string
  first_name: string
  last_name: string
  is_sysadmin: boolean
  chapter_count: number
  last_login_at?: string
  created_at: string
}

export interface AuditEntry {
  id: string
  actor_id?: string
  actor_email?: string
  action: string
  entity_type?: string
  entity_id?: string
  ip_address?: string
  created_at: string
}

export interface PlatformListResponse<T> {
  data: T[]
  meta: { total: number }
}

export async function sysListChapters(params?: {
  page?: number
  per_page?: number
}): Promise<PlatformListResponse<ChapterSummary>> {
  const { data } = await apiClient.get<PlatformListResponse<ChapterSummary>>('/sys/chapters', { params })
  return data
}

export async function sysCreateChapter(payload: {
  name: string
  greek_letters: string
  city: string
  state_code: string
  university?: string
}): Promise<ChapterSummary> {
  const { data } = await apiClient.post<APIResponse<ChapterSummary>>('/sys/chapters', payload)
  return data.data
}

export async function sysGetChapter(id: string): Promise<ChapterSummary> {
  const { data } = await apiClient.get<APIResponse<ChapterSummary>>(`/sys/chapters/${id}`)
  return data.data
}

export async function sysUpdateSubscription(
  id: string,
  payload: { subscription_status: string; plan_tier: string }
): Promise<ChapterSummary> {
  const { data } = await apiClient.put<APIResponse<ChapterSummary>>(`/sys/chapters/${id}/subscription`, payload)
  return data.data
}

export async function sysDeleteChapter(id: string): Promise<void> {
  await apiClient.delete(`/sys/chapters/${id}`)
}

export async function sysListUsers(params?: {
  page?: number
  per_page?: number
}): Promise<PlatformListResponse<SysUser>> {
  const { data } = await apiClient.get<PlatformListResponse<SysUser>>('/sys/users', { params })
  return data
}

export async function sysGetAuditLog(params?: {
  page?: number
  per_page?: number
}): Promise<PlatformListResponse<AuditEntry>> {
  const { data } = await apiClient.get<PlatformListResponse<AuditEntry>>('/sys/audit-log', { params })
  return data
}

export async function sysRecalculateXP(): Promise<{ members_updated: number }> {
  const { data } = await apiClient.post<APIResponse<{ members_updated: number }>>('/sys/xp/recalculate-all', {})
  return data.data
}

export async function getBillingPortalUrl(customerId: string): Promise<string> {
  const { data } = await apiClient.post<APIResponse<{ url: string }>>('/settings/billing/portal', {
    customer_id: customerId,
  })
  return data.data.url
}
