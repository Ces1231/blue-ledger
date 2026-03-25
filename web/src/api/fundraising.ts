import apiClient from './client'
import type { APIResponse } from '../types'

export interface Campaign {
  id: string
  title: string
  description: string
  goal_cents: number
  raised_cents: number
  deadline?: string
  active: boolean
  created_at: string
}

export interface Contribution {
  id: string
  campaign_id: string
  member_id: string
  amount_cents: number
  note?: string
  created_at: string
}

export async function getCampaigns(): Promise<Campaign[]> {
  const { data } = await apiClient.get<APIResponse<Campaign[]>>('/fundraising')
  return data.data ?? []
}

export async function getCampaign(id: string): Promise<Campaign> {
  const { data } = await apiClient.get<APIResponse<Campaign>>(`/fundraising/${id}`)
  return data.data
}

export async function createCampaign(input: Partial<Campaign>): Promise<Campaign> {
  const { data } = await apiClient.post<APIResponse<Campaign>>('/fundraising', input)
  return data.data
}

export async function contribute(campaignId: string, amountCents: number, note?: string): Promise<Contribution> {
  const { data } = await apiClient.post<APIResponse<Contribution>>(`/fundraising/${campaignId}/contribute`, {
    amount_cents: amountCents,
    note,
  })
  return data.data
}
