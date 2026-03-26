import apiClient from './client'
import type { Announcement, APIResponse } from '../types'

export async function getAnnouncements(): Promise<Announcement[]> {
  const { data } = await apiClient.get<APIResponse<Announcement[]>>('/announcements')
  return data.data
}

export interface CreateAnnouncementPayload {
  title: string
  body: string
  category?: string
  is_pinned?: boolean
}

export async function createAnnouncement(payload: CreateAnnouncementPayload): Promise<Announcement> {
  const { data } = await apiClient.post<APIResponse<Announcement>>('/announcements', payload)
  return data.data
}

export async function updateAnnouncement(
  id: string,
  payload: Partial<CreateAnnouncementPayload>
): Promise<Announcement> {
  const { data } = await apiClient.put<APIResponse<Announcement>>(`/announcements/${id}`, payload)
  return data.data
}

export async function deleteAnnouncement(id: string): Promise<void> {
  await apiClient.delete(`/announcements/${id}`)
}

export async function pinAnnouncement(id: string, pinned: boolean): Promise<Announcement> {
  const { data } = await apiClient.put<APIResponse<Announcement>>(
    `/announcements/${id}/pin`,
    { pinned }
  )
  return data.data
}
