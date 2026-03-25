import apiClient from './client'
import type { APIResponse } from '../types'

export interface AIConfig {
  id?: string
  provider: 'claude' | 'openai'
  model: string
  system_prompt: string
  enabled: boolean
}

export interface AIMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  provider: string
  model: string
  created_at: string
}

export interface ChatMessage {
  role: 'user' | 'assistant' | 'system'
  content: string
}

export async function getAIConfig(): Promise<AIConfig> {
  const { data } = await apiClient.get<APIResponse<AIConfig>>('/ai/config')
  return data.data
}

export async function updateAIConfig(input: Partial<AIConfig>): Promise<AIConfig> {
  const { data } = await apiClient.put<APIResponse<AIConfig>>('/ai/config', input)
  return data.data
}

export async function getAIChatHistory(): Promise<AIMessage[]> {
  const { data } = await apiClient.get<APIResponse<AIMessage[]>>('/ai/history')
  return data.data ?? []
}

/**
 * Streams a chat response from /ai/chat using fetch + ReadableStream.
 * Calls onDelta for each text chunk, onDone when complete.
 */
export async function streamChat(
  message: string,
  history: ChatMessage[],
  onDelta: (text: string) => void,
  onDone: () => void,
  onError: (err: string) => void
): Promise<void> {
  const token = localStorage.getItem('access_token')
  const response = await fetch('/api/v1/ai/chat', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: token ? `Bearer ${token}` : '',
    },
    body: JSON.stringify({ message, history }),
  })

  if (!response.ok) {
    onError(`Request failed: ${response.status}`)
    return
  }

  const reader = response.body?.getReader()
  if (!reader) {
    onError('Streaming not supported')
    return
  }

  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() ?? ''

    for (const line of lines) {
      if (!line.startsWith('data:')) continue
      const data = line.slice(5).trim()
      if (data === '[DONE]') {
        onDone()
        return
      }
      try {
        const parsed = JSON.parse(data)
        if (parsed.error) {
          onError(parsed.error)
          return
        }
        if (parsed.delta) {
          onDelta(parsed.delta)
        }
      } catch {
        // skip malformed
      }
    }
  }

  onDone()
}
