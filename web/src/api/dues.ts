import apiClient from './client'
import type { DuesRecord, PaginatedResponse, APIResponse } from '../types'

export async function getDues(params?: { semester?: string }): Promise<PaginatedResponse<DuesRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<DuesRecord>>('/dues', { params })
  return data
}

export async function getMyDues(): Promise<DuesRecord[]> {
  const { data } = await apiClient.get<APIResponse<DuesRecord[]>>('/dues/me')
  return data.data
}

export interface CreateDuesPayload {
  member_id: string
  semester: string
  amount_cents: number
  due_date: string
}

export async function createDues(payload: CreateDuesPayload): Promise<DuesRecord> {
  const { data } = await apiClient.post<APIResponse<DuesRecord>>('/dues', payload)
  return data.data
}

export async function markDuesPaid(id: string, paymentMethod: string): Promise<DuesRecord> {
  const { data } = await apiClient.put<APIResponse<DuesRecord>>(`/dues/${id}/mark-paid`, {
    payment_method: paymentMethod,
  })
  return data.data
}
