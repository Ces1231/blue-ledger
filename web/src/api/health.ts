import apiClient from './client'
import type { APIResponse } from '../types'

export interface ChapterHealth {
  attendance_rate: number
  avg_xp: number
  total_service_hours: number
  dues_paid_rate: number
  active_members: number
  total_members: number
}

export interface AttendanceStat {
  event_name: string
  event_date: string
  total_rsvps: number
  total_attended: number
  rate: number
}

export interface XPStat {
  level: string
  count: number
  avg_xp: number
}

export interface ServiceStat {
  period: string
  hours: number
  count: number
}

export interface DuesStat {
  status: string
  count: number
}

export async function getHealthOverview(): Promise<ChapterHealth> {
  const { data } = await apiClient.get<APIResponse<ChapterHealth>>('/health')
  return data.data
}

export async function getAttendanceStats(limit?: number): Promise<AttendanceStat[]> {
  const { data } = await apiClient.get<APIResponse<AttendanceStat[]>>('/health/attendance', {
    params: { limit },
  })
  return data.data
}

export async function getXPDistribution(): Promise<XPStat[]> {
  const { data } = await apiClient.get<APIResponse<XPStat[]>>('/health/xp')
  return data.data
}

export async function getServiceStats(months?: number): Promise<ServiceStat[]> {
  const { data } = await apiClient.get<APIResponse<ServiceStat[]>>('/health/service', {
    params: { months },
  })
  return data.data
}

export async function getDuesBreakdown(): Promise<DuesStat[]> {
  const { data } = await apiClient.get<APIResponse<DuesStat[]>>('/health/dues')
  return data.data
}
