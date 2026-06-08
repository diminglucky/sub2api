import { apiClient } from '../client'
import type { BasePaginationResponse } from '@/types'
import type { LotteryEvent } from '../lottery'

export interface AdminLotteryPrizeRequest {
  type: 'balance' | 'card'
  name: string
  quantity: number
  amount?: number
  card_content?: string
  sort_order?: number
}

export interface CreateLotteryRequest {
  title: string
  description?: string
  status: 'draft' | 'active' | 'archived'
  starts_at?: number | null
  draw_at: number
  prizes: AdminLotteryPrizeRequest[]
}

export interface UpdateLotteryRequest {
  title?: string
  description?: string
  status?: 'draft' | 'active' | 'drawn' | 'archived'
  starts_at?: number | null
  draw_at?: number
  prizes?: AdminLotteryPrizeRequest[]
}

export async function list(
  page = 1,
  pageSize = 20,
  filters?: { status?: string; search?: string; sort_by?: string; sort_order?: 'asc' | 'desc' }
): Promise<BasePaginationResponse<LotteryEvent>> {
  const { data } = await apiClient.get<BasePaginationResponse<LotteryEvent>>('/admin/lotteries', {
    params: { page, page_size: pageSize, ...filters }
  })
  return data
}

export async function getById(id: number): Promise<LotteryEvent> {
  const { data } = await apiClient.get<LotteryEvent>(`/admin/lotteries/${id}`)
  return data
}

export async function create(request: CreateLotteryRequest): Promise<LotteryEvent> {
  const { data } = await apiClient.post<LotteryEvent>('/admin/lotteries', request)
  return data
}

export async function update(id: number, request: UpdateLotteryRequest): Promise<LotteryEvent> {
  const { data } = await apiClient.put<LotteryEvent>(`/admin/lotteries/${id}`, request)
  return data
}

export async function deleteLottery(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/lotteries/${id}`)
  return data
}

export async function draw(id: number): Promise<LotteryEvent> {
  const { data } = await apiClient.post<LotteryEvent>(`/admin/lotteries/${id}/draw`)
  return data
}

const lotteryAdminAPI = {
  list,
  getById,
  create,
  update,
  delete: deleteLottery,
  draw
}

export default lotteryAdminAPI
