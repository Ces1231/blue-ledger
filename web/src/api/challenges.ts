import apiClient from './client'

export interface Challenge {
  id: string
  chapter_id: string
  challenger_id: string
  challenged_id: string
  type: 'xp_duel' | 'service_race' | 'trivia' | 'streak_showdown'
  status: 'pending' | 'accepted' | 'declined' | 'active' | 'completed' | 'expired'
  xp_stake: number
  game_data: Record<string, unknown>
  winner_id?: string
  expires_at: string
  accepted_at?: string
  completed_at?: string
  created_at: string
  challenger_name?: string
  challenged_name?: string
  challenger_level?: string
  challenged_level?: string
}

export interface SendChallengeInput {
  challenged_id: string
  type: Challenge['type']
  xp_stake: number
}

export async function listChallenges(): Promise<Challenge[]> {
  const res = await apiClient.get<{ data: Challenge[] }>('/v1/challenges')
  return res.data.data ?? []
}

export async function getChallenge(id: string): Promise<Challenge> {
  const res = await apiClient.get<{ data: Challenge }>(`/v1/challenges/${id}`)
  return res.data.data
}

export async function sendChallenge(input: SendChallengeInput): Promise<Challenge> {
  const res = await apiClient.post<{ data: Challenge }>('/v1/challenges', input)
  return res.data.data
}

export async function acceptChallenge(id: string): Promise<Challenge> {
  const res = await apiClient.post<{ data: Challenge }>(`/v1/challenges/${id}/accept`)
  return res.data.data
}

export async function declineChallenge(id: string): Promise<Challenge> {
  const res = await apiClient.post<{ data: Challenge }>(`/v1/challenges/${id}/decline`)
  return res.data.data
}

export async function submitAnswers(id: string, answers: number[]): Promise<Challenge> {
  const res = await apiClient.post<{ data: Challenge }>(`/v1/challenges/${id}/submit`, { answers })
  return res.data.data
}
