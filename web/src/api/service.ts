import apiClient from './client'
import type { ServiceLogEntry, APIResponse } from '../types'

export interface ServiceListResponse {
  data: ServiceLogEntry[]
  meta: { total: number }
}

export async function getServiceLog(params?: {
  page?: number
  per_page?: number
}): Promise<ServiceListResponse> {
  const { data } = await apiClient.get<ServiceListResponse>('/service', { params })
  return data
}

export interface LogServicePayload {
  event_name: string
  organization?: string
  service_date: string
  hours: number
  notes?: string
}

export async function logService(payload: LogServicePayload): Promise<ServiceLogEntry> {
  const { data } = await apiClient.post<APIResponse<ServiceLogEntry>>('/service', payload)
  return data.data
}

export async function verifyService(id: string): Promise<ServiceLogEntry> {
  const { data } = await apiClient.put<APIResponse<ServiceLogEntry>>(`/service/${id}/verify`, {})
  return data.data
}

export async function getMemberServiceLog(
  memberId: string,
  params?: { page?: number; per_page?: number }
): Promise<ServiceListResponse> {
  const { data } = await apiClient.get<ServiceListResponse>(`/service/member/${memberId}`, { params })
  return data
}
