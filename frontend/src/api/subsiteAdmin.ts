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
  image_price_1k: number | null
  image_price_2k: number | null
  image_price_4k: number | null
  status: string
}

export interface SubsitePriceInput {
  scope: 'group' | 'model'
  group_id?: number | null
  model?: string | null
  rate_multiplier: number
  image_price_1k?: number | null
  image_price_2k?: number | null
  image_price_4k?: number | null
  status?: string
}

export interface SubsiteChannel {
  group_id: number
  name: string
  platform: string
  main_rate_multiplier: number
  subsite_rate_multiplier: number | null
  main_image_price_1k: number | null
  main_image_price_2k: number | null
  main_image_price_4k: number | null
  subsite_image_price_1k: number | null
  subsite_image_price_2k: number | null
  subsite_image_price_4k: number | null
  enabled: boolean
}

export interface SubsiteSettings {
  slug: string
  domain: string
  name: string
  logo_url: string
  theme_color: string
  status: string
}

export interface SubsiteSettingsInput {
  name: string
  logo_url: string
  theme_color: string
}

export interface SubsiteMember {
  user_id: number
  email: string
  role: string
  source: string
  balance: number
  usage_count: number
  usage_cost: number
  recharge_count: number
  recharge_amount: number
  joined_at: string
}

export interface SubsiteMemberPage {
  items: SubsiteMember[]
  page: number
  page_size: number
}

export interface SubsiteModelPrice {
  group_id: number
  group_name: string
  main_multiplier: number
  model: string
  platform: string
  billing_mode: string
  subsite_multiplier: number | null
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

export async function getSubsiteSettings(): Promise<SubsiteSettings> {
  const { data } = await apiClient.get<SubsiteSettings>(
    subsiteURL('/api/internal/downstream/v1/admin/settings')
  )
  return data
}

export async function listSubsiteUsers(page = 1, pageSize = 50): Promise<SubsiteMemberPage> {
  const { data } = await apiClient.get<SubsiteMemberPage>(
    subsiteURL(`/api/internal/downstream/v1/admin/users?page=${page}&page_size=${pageSize}`)
  )
  return data
}

export async function listSubsiteModels(): Promise<SubsiteModelPrice[]> {
  const { data } = await apiClient.get<SubsiteModelPrice[]>(
    subsiteURL('/api/internal/downstream/v1/admin/models')
  )
  return data
}

export async function updateSubsiteSettings(input: SubsiteSettingsInput): Promise<SubsiteSettings> {
  const { data } = await apiClient.put<SubsiteSettings>(
    subsiteURL('/api/internal/downstream/v1/admin/settings'),
    input
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
  getSettings: getSubsiteSettings,
  updateSettings: updateSubsiteSettings,
  listUsers: listSubsiteUsers,
  listModels: listSubsiteModels,
  listChannels: listSubsiteChannels,
  listPrices: listSubsitePrices,
  createPrice: createSubsitePrice,
  updatePrice: updateSubsitePrice,
  deletePrice: deleteSubsitePrice
}

export default subsiteAdminAPI
