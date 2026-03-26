import apiClient from './client'
import type { APIResponse } from '../types'

export interface JobPosting {
  id: string
  title: string
  company: string
  location?: string
  description: string
  url?: string
  posted_by: string
  active: boolean
  created_at: string
}

export async function getJobs(): Promise<JobPosting[]> {
  const { data } = await apiClient.get<APIResponse<JobPosting[]>>('/job-board')
  return data.data ?? []
}

export async function createJob(input: Partial<JobPosting>): Promise<JobPosting> {
  const { data } = await apiClient.post<APIResponse<JobPosting>>('/job-board', input)
  return data.data
}

export async function deleteJob(id: string): Promise<void> {
  await apiClient.delete(`/job-board/${id}`)
}
