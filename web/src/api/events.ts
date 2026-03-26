import apiClient from './client'
import type { Event, PaginatedResponse, APIResponse } from '../types'

export async function getEvents(params?: {
  page?: number
  per_page?: number
  upcoming?: boolean
}): Promise<PaginatedResponse<Event>> {
  const { data } = await apiClient.get<PaginatedResponse<Event>>('/events', { params })
  return data
}

export async function getEvent(id: string): Promise<Event> {
  const { data } = await apiClient.get<APIResponse<Event>>(`/events/${id}`)
  return data.data
}

export interface CreateEventPayload {
  name: string
  description?: string
  event_date: string
  event_time?: string
  location?: string
  event_type: string
  xp_attend?: number
  xp_rsvp?: number
  rsvp_deadline?: string
  capacity?: number
}

export async function createEvent(payload: CreateEventPayload): Promise<Event> {
  const { data } = await apiClient.post<APIResponse<Event>>('/events', payload)
  return data.data
}

export async function updateEvent(id: string, payload: Partial<CreateEventPayload>): Promise<Event> {
  const { data } = await apiClient.put<APIResponse<Event>>(`/events/${id}`, payload)
  return data.data
}

export async function deleteEvent(id: string): Promise<void> {
  await apiClient.delete(`/events/${id}`)
}

export async function rsvpEvent(id: string, status: 'yes' | 'no' | 'maybe'): Promise<void> {
  await apiClient.post(`/events/${id}/rsvp`, { status })
}

export async function getQRToken(eventId: string): Promise<string> {
  const { data } = await apiClient.get<APIResponse<{ token: string }>>(`/events/${eventId}/qr-token`)
  return data.data.token
}

export interface CheckInQRPayload {
  member_display_id: string
  scanner_token: string
}

export interface CheckInResult {
  member_id: string
  member_name: string
  xp_awarded: number
  level: string
  new_badges: string[]
}

export async function checkinQR(eventId: string, payload: CheckInQRPayload): Promise<CheckInResult> {
  const { data } = await apiClient.post<APIResponse<CheckInResult>>(
    `/events/${eventId}/checkin/qr`,
    payload
  )
  return data.data
}
