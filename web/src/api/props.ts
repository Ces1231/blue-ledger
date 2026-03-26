import apiClient from './client'
import type { Props, APIResponse } from '../types'

export interface PropsListResponse {
  data: Props[]
  meta: { total: number }
}

export async function getProps(params?: { page?: number; per_page?: number }): Promise<PropsListResponse> {
  const { data } = await apiClient.get<PropsListResponse>('/props', { params })
  return data
}

export interface GivePropsPayload {
  to_member_id: string
  category: string
  message?: string
}

export async function giveProps(payload: GivePropsPayload): Promise<Props> {
  const { data } = await apiClient.post<APIResponse<Props>>('/props', payload)
  return data.data
}

export async function getReceivedProps(params?: { page?: number; per_page?: number }): Promise<PropsListResponse> {
  const { data } = await apiClient.get<PropsListResponse>('/props/received', { params })
  return data
}
