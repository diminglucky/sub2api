/**
 * Sub-site-scoped admin API.
 *
 * These endpoints live under /api/internal/downstream/v1 and are authorized by
 * the caller's membership role in the current sub-site, not by the main-site
 * admin role. A sub-site owner can therefore manage only its own prices.
 */

import { apiClient } from './client'

export interface SubsiteAdminSummary {
  subsite_id: number
  member_count: number
  recharge_count: number
  recharge_amount: number
  usage_count: number
  usage_cost: number
  settlement_pending: number
}

export interface SubsitePriceOverride {
  id: number
  subsite_id: number
  scope: 'group' | 'model'
  group_id: number | null
  model: string | null
  rate_multiplier: number
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_read_price: number | null
  per_request_price: number | null
  status: string
}

export interface SubsitePriceInput {
  scope: 'group' | 'model'
  group_id?: number | null
  model?: string | null
  rate_multiplier: number
  status?: string
}

export interface SubsiteChannel {
  group_id: number
  name: string
  platform: string
  main_rate_multiplier: number
  subsite_rate_multiplier: number | null
  enabled: boolean
}

// The shared apiClient prefixes /api/v1, so sub-site endpoints use absolute
// URLs to bypass that base and hit /api/internal/... directly.
function subsiteURL(path: string): string {
  if (typeof window === 'undefined') return path
  return `${window.location.origin}${path}`
}

export async function getSubsiteAdminSummary(): Promise<SubsiteAdminSummary> {
  const { data } = await apiClient.get<SubsiteAdminSummary>(
    subsiteURL('/api/internal/downstream/v1/admin/summary')
  )
  return data
}

export async function listSubsitePrices(): Promise<SubsitePriceOverride[]> {
  const { data } = await apiClient.get<SubsitePriceOverride[]>(
    subsiteURL('/api/internal/downstream/v1/admin/prices')
  )
  return data
}

export async function listSubsiteChannels(): Promise<SubsiteChannel[]> {
  const { data } = await apiClient.get<SubsiteChannel[]>(
    subsiteURL('/api/internal/downstream/v1/admin/channels')
  )
  return data
}

export async function createSubsitePrice(input: SubsitePriceInput): Promise<SubsitePriceOverride> {
  const { data } = await apiClient.post<SubsitePriceOverride>(
    subsiteURL('/api/internal/downstream/v1/admin/prices'),
    input
  )
  return data
}

export async function updateSubsitePrice(
  id: number,
  input: SubsitePriceInput
): Promise<SubsitePriceOverride> {
  const { data } = await apiClient.put<SubsitePriceOverride>(
    subsiteURL(`/api/internal/downstream/v1/admin/prices/${id}`),
    input
  )
  return data
}

export async function deleteSubsitePrice(id: number): Promise<void> {
  await apiClient.delete(subsiteURL(`/api/internal/downstream/v1/admin/prices/${id}`))
}

export const subsiteAdminAPI = {
  getSummary: getSubsiteAdminSummary,
  listChannels: listSubsiteChannels,
  listPrices: listSubsitePrices,
  createPrice: createSubsitePrice,
  updatePrice: updateSubsitePrice,
  deletePrice: deleteSubsitePrice
}

export default subsiteAdminAPI
