import { apiClient } from './client'

export interface LotteryPrize {
  id: number
  type: 'balance' | 'card'
  name: string
  quantity: number
  amount?: number
  card_content?: string
  sort_order: number
}

export interface LotteryWinner {
  id: number
  event_id: number
  prize_id: number
  user_id: number
  user_email?: string
  user_display_name?: string
  prize_type: 'balance' | 'card'
  prize_name: string
  amount?: number
  card_content?: string
  delivered_at: number
}

export interface LotteryEvent {
  id: number
  title: string
  description: string
  status: 'draft' | 'active' | 'drawn' | 'archived'
  starts_at?: number
  draw_at: number
  drawn_at?: number
  created_at: number
  updated_at: number
  prizes: LotteryPrize[]
  entry_count: number
  winners?: LotteryWinner[]
  joined: boolean
  my_joined_at?: number
  my_winner?: LotteryWinner
}

export async function list(): Promise<LotteryEvent[]> {
  const { data } = await apiClient.get<LotteryEvent[]>('/lotteries')
  return data
}

export async function join(id: number): Promise<LotteryEvent> {
  const { data } = await apiClient.post<LotteryEvent>(`/lotteries/${id}/join`)
  return data
}

export const lotteryAPI = {
  list,
  join
}

export default lotteryAPI
