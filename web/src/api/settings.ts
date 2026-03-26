import apiClient from './client'
import type { APIResponse } from '../types'

export interface ChapterSettings {
  zeffy_form_id?: string
  notification_prefs?: Record<string, boolean>
  allow_public_directory: boolean
  dark_mode: boolean
  semester_label?: string
  xp_levels?: Array<{ key: string; label: string; min: number }>
}

export interface PointEconomy {
  attend_event: number
  rsvp_event: number
  service_per_hour: number
  props: number
  dues_paid: number
  badge_bonus: number
}

export async function getSettings(): Promise<ChapterSettings> {
  const { data } = await apiClient.get<APIResponse<ChapterSettings>>('/settings')
  return data.data
}

export interface UpdateSettingsPayload {
  allow_public_directory?: boolean
  dark_mode?: boolean
  notification_prefs?: Record<string, boolean>
  semester_label?: string
}

export async function updateSettings(payload: UpdateSettingsPayload): Promise<ChapterSettings> {
  const { data } = await apiClient.put<APIResponse<ChapterSettings>>('/settings', payload)
  return data.data
}

export async function getPointEconomy(): Promise<PointEconomy> {
  const { data } = await apiClient.get<APIResponse<PointEconomy>>('/settings/point-economy')
  return data.data
}

export async function updatePointEconomy(payload: PointEconomy): Promise<PointEconomy> {
  const { data } = await apiClient.put<APIResponse<PointEconomy>>('/settings/point-economy', payload)
  return data.data
}

export async function saveZeffyConfig(formId: string): Promise<ChapterSettings> {
  const { data } = await apiClient.post<APIResponse<ChapterSettings>>('/settings/zeffy-config', { form_id: formId })
  return data.data
}
