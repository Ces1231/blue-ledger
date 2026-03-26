import apiClient from './client'
import type { AuthResponse, APIResponse, UserSummary } from '../types'

export interface LoginPayload {
  email: string
  password: string
}

export interface RegisterPayload {
  chapter_name: string
  greek_letters: string
  city: string
  state_code: string
  university?: string
  first_name: string
  last_name: string
  email: string
  password: string
}

export async function login(payload: LoginPayload): Promise<AuthResponse> {
  const { data } = await apiClient.post<APIResponse<AuthResponse>>('/auth/login', payload)
  return data.data
}

export async function register(payload: RegisterPayload): Promise<AuthResponse> {
  const { data } = await apiClient.post<APIResponse<AuthResponse>>('/auth/register', payload)
  return data.data
}

export async function requestMagicLink(email: string): Promise<void> {
  await apiClient.post('/auth/magic-link', { email })
}

export async function consumeMagicLink(token: string): Promise<AuthResponse> {
  const { data } = await apiClient.get<APIResponse<AuthResponse>>(`/auth/magic?token=${token}`)
  return data.data
}

export async function refreshTokens(refreshToken: string): Promise<AuthResponse> {
  const { data } = await apiClient.post<APIResponse<AuthResponse>>('/auth/refresh', {
    refresh_token: refreshToken,
  })
  return data.data
}

export async function logout(refreshToken?: string): Promise<void> {
  try {
    await apiClient.post('/auth/logout', { refresh_token: refreshToken })
  } catch {
    // Ignore logout errors — always clear local state
  }
}

export async function getMe(): Promise<UserSummary> {
  const { data } = await apiClient.get<APIResponse<UserSummary>>('/auth/me')
  return data.data
}
