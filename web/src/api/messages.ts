import apiClient from './client'
import type { APIResponse } from '../types'

export interface MessageThread {
  id: string
  chapter_id: string
  subject: string
  created_by: string
  is_group_chat: boolean
  created_at: string
  last_message_at?: string
  last_message_body?: string
  message_count: number
  author_first_name?: string
  author_last_name?: string
}

export interface Message {
  id: string
  thread_id: string
  chapter_id: string
  sender_id: string
  body: string
  sent_at: string
  sender_first_name?: string
  sender_last_name?: string
  sender_avatar_url?: string
}

export async function getThreads(): Promise<MessageThread[]> {
  const { data } = await apiClient.get<APIResponse<MessageThread[]>>('/messages/threads')
  return data.data
}

export interface CreateThreadPayload {
  subject: string
  is_group_chat?: boolean
  participant_ids?: string[]
  initial_message: string
}

export async function createThread(payload: CreateThreadPayload): Promise<MessageThread> {
  const { data } = await apiClient.post<APIResponse<MessageThread>>('/messages/threads', payload)
  return data.data
}

export async function getThread(threadId: string): Promise<Message[]> {
  const { data } = await apiClient.get<APIResponse<Message[]>>(`/messages/threads/${threadId}`)
  return data.data
}

export async function sendMessage(threadId: string, body: string): Promise<Message> {
  const { data } = await apiClient.post<APIResponse<Message>>(`/messages/threads/${threadId}/messages`, { body })
  return data.data
}
