import apiClient from './client'
import type { APIResponse } from '../types'

export interface Notification {
  id: string
  member_id: string
  type: string
  title: string
  body: string
  read: boolean
  created_at: string
}

export async function getNotifications(): Promise<Notification[]> {
  const { data } = await apiClient.get<APIResponse<Notification[]>>('/notifications')
  return data.data ?? []
}

export async function markRead(id: string): Promise<void> {
  await apiClient.put(`/notifications/${id}/read`)
}

export async function markAllRead(): Promise<void> {
  await apiClient.put('/notifications/read-all')
}
