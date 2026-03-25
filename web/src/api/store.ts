import apiClient from './client'
import type { APIResponse } from '../types'

export interface StoreItem {
  id: string
  chapter_id: string
  name: string
  description?: string
  image_url?: string
  xp_cost: number
  quantity?: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface StoreOrder {
  id: string
  chapter_id: string
  member_id: string
  item_id: string
  xp_spent: number
  status: string
  created_at: string
  item_name?: string
  member_first_name?: string
  member_last_name?: string
}

export interface OrdersListResponse {
  data: StoreOrder[]
  meta: { total: number }
}

export async function getStoreItems(): Promise<StoreItem[]> {
  const { data } = await apiClient.get<APIResponse<StoreItem[]>>('/store')
  return data.data
}

export interface CreateStoreItemPayload {
  name: string
  description?: string
  image_url?: string
  xp_cost: number
  quantity?: number
}

export async function createStoreItem(payload: CreateStoreItemPayload): Promise<StoreItem> {
  const { data } = await apiClient.post<APIResponse<StoreItem>>('/store', payload)
  return data.data
}

export async function purchaseItem(itemId: string): Promise<StoreOrder> {
  const { data } = await apiClient.post<APIResponse<StoreOrder>>(`/store/${itemId}/purchase`, {})
  return data.data
}

export async function getOrders(params?: {
  page?: number
  per_page?: number
}): Promise<OrdersListResponse> {
  const { data } = await apiClient.get<OrdersListResponse>('/store/orders', { params })
  return data
}
