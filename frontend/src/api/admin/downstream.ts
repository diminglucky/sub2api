/**
 * Admin downstream subsite API
 * Provisioning for branded downstream sites and their price overrides.
 */

import { apiClient } from '../client'

export interface DownstreamSubsite {
  id: number
  slug: string
  domain: string
  name: string
  logo_url: string
  theme_color: string
  status: string
  admin_user_id: number | null
  created_at: string
  updated_at: string
}

export interface DownstreamSubsiteInput {
  slug: string
  domain: string
  name: string
  logo_url?: string
  theme_color?: string
  status?: string
  admin_user_id?: number | null
}

export interface DownstreamPriceOverride {
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
  created_at: string
  updated_at: string
}

export interface DownstreamPriceOverrideInput {
  scope: 'group' | 'model'
  group_id?: number | null
  model?: string | null
  rate_multiplier: number
  status?: string
}

export async function listSubsites(): Promise<DownstreamSubsite[]> {
  const { data } = await apiClient.get<DownstreamSubsite[]>('/admin/downstream/subsites')
  return data
}

export async function createSubsite(input: DownstreamSubsiteInput): Promise<DownstreamSubsite> {
  const { data } = await apiClient.post<DownstreamSubsite>('/admin/downstream/subsites', input)
  return data
}

export async function updateSubsite(
  id: number,
  input: DownstreamSubsiteInput
): Promise<DownstreamSubsite> {
  const { data } = await apiClient.put<DownstreamSubsite>(`/admin/downstream/subsites/${id}`, input)
  return data
}

export async function disableSubsite(id: number): Promise<DownstreamSubsite> {
  const { data } = await apiClient.post<DownstreamSubsite>(
    `/admin/downstream/subsites/${id}/disable`
  )
  return data
}

export async function listPrices(subsiteId: number): Promise<DownstreamPriceOverride[]> {
  const { data } = await apiClient.get<DownstreamPriceOverride[]>(
    `/admin/downstream/subsites/${subsiteId}/prices`
  )
  return data
}

export async function createPrice(
  subsiteId: number,
  input: DownstreamPriceOverrideInput
): Promise<DownstreamPriceOverride> {
  const { data } = await apiClient.post<DownstreamPriceOverride>(
    `/admin/downstream/subsites/${subsiteId}/prices`,
    input
  )
  return data
}

export async function updatePrice(
  subsiteId: number,
  priceId: number,
  input: DownstreamPriceOverrideInput
): Promise<DownstreamPriceOverride> {
  const { data } = await apiClient.put<DownstreamPriceOverride>(
    `/admin/downstream/subsites/${subsiteId}/prices/${priceId}`,
    input
  )
  return data
}

export async function deletePrice(subsiteId: number, priceId: number): Promise<void> {
  await apiClient.delete(`/admin/downstream/subsites/${subsiteId}/prices/${priceId}`)
}

export const downstreamAdminAPI = {
  listSubsites,
  createSubsite,
  updateSubsite,
  disableSubsite,
  listPrices,
  createPrice,
  updatePrice,
  deletePrice
}

export default downstreamAdminAPI
